/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package node

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/projectriff/riff-buildpack/pkg/invoker"
)

const (
	// Dependency is a key identifying the node invoker dependency in the build plan.
	Dependency = "riff-invoker-node"

	// functionArtifact is a key identifying the path to the function entrypoint in the build plan.
	FunctionArtifact = "fn"
)

// RiffNodeInvoker represents the Node invoker contributed by the buildpack.
type RiffNodeInvoker struct {
	// A reference to the user function source tree.
	application application.Application

	// The file in the function tree that is the entrypoint.
	// May be empty, in which case the function is require()d as a node module.
	functionJS string

	// Provides access to the launch layers, used to craft the process commands.
	layers layers.Layers

	// A dedicated layer for the node invoker itself. Cacheable once npm-installed
	invokerLayer layers.DependencyLayer

	// A dedicated layer for the function location. Not cacheable, as it changes with the value of functionJS.
	functionLayer layers.Layer
}

// Contribute expands the node invoker tgz and creates launch configurations that run "node server.js"
func (r RiffNodeInvoker) Contribute() error {
	if err := r.invokerLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		if e := helper.ExtractTarGz(artifact, layer.Root, 1); e != nil {
			return e
		}
		layer.Logger.SubsequentLine("npm-installing the node invoker")
		cmd := exec.Command("npm", "install", "--production")
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.Dir = layer.Root
		if e := cmd.Run(); e != nil {
			return e
		}

		if e := layer.OverrideLaunchEnv("HOST", "0.0.0.0"); e != nil {
			return e
		}
		if e := layer.OverrideLaunchEnv("HTTP_PORT", "8080"); e != nil {
			return e
		}

		return nil
	}, layers.Launch); err != nil {
		return err
	}

	if err := r.functionLayer.Contribute(marker{"NodeJS", r.functionJS}, func(layer layers.Layer) error {
		return layer.OverrideLaunchEnv("FUNCTION_URI", filepath.Join(r.application.Root, r.functionJS))
	}, layers.Launch); err != nil {
		return err
	}

	command := fmt.Sprintf(`node %s/server.js`, r.invokerLayer.Root)

	return r.layers.WriteMetadata(layers.Metadata{
		Processes: layers.Processes{
			layers.Process{Type: "web", Command: command},
			layers.Process{Type: "function", Command: command},
		},
	})
}

func NewNodeInvoker(build build.Build) (invoker.RiffInvoker, bool, error) {
	bp, ok := build.BuildPlan[Dependency]
	if !ok {
		return RiffNodeInvoker{}, false, nil
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return RiffNodeInvoker{}, false, err
	}

	dep, err := deps.Best(Dependency, bp.Version, build.Stack)
	if err != nil {
		return RiffNodeInvoker{}, false, err
	}

	functionJS, ok := bp.Metadata[FunctionArtifact].(string)
	if !ok {
		return RiffNodeInvoker{}, false, fmt.Errorf("node metadata of incorrect type: %v", bp.Metadata[FunctionArtifact])
	}

	return RiffNodeInvoker{
		application:   build.Application,
		functionJS:    functionJS,
		layers:        build.Layers,
		invokerLayer:  build.Layers.DependencyLayer(dep),
		functionLayer: build.Layers.Layer("function"),
	}, true, nil

}

type marker struct {
	Language string `toml:"language"`
	Function string `toml:"function"`
}

func (m marker) Identity() (string, string) {
	return m.Language, m.Function
}
