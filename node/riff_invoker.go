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
	. "github.com/projectriff/riff-buildpack/plugins"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/nodejs-cnb/node"
	"github.com/projectriff/riff-buildpack/metadata"
)

const (
	// Dependency is a key identifying the node invoker dependency in the build plan.
	Dependency = "riff-invoker-node"

	// functionArtifact is a key identifying the path to the function entrypoint in the build plan.
	FunctionArtifact = "fn"
)

func BuildPlanContribution(detect detect.Detect, metadata metadata.Metadata) buildplan.BuildPlan {
	n := detect.BuildPlan[node.Dependency]
	if n.Metadata == nil {
		n.Metadata = buildplan.Metadata{}
	}
	n.Metadata["launch"] = true
	n.Metadata["build"] = true

	r := detect.BuildPlan[Dependency]
	if r.Metadata == nil {
		r.Metadata = buildplan.Metadata{}
	}
	r.Metadata[FunctionArtifact] = metadata.Artifact

	return buildplan.BuildPlan{node.Dependency: n, Dependency: r}
}

// Contribute expands the node invoker tgz and creates launch configurations that run "node server.js"
func Contribute(r RiffInvoker) error {
	if err := r.InvokerLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
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

	if err := r.FunctionLayer.Contribute(marker{"NodeJS", r.Handler}, func(layer layers.Layer) error {
		return layer.OverrideLaunchEnv("FUNCTION_URI", filepath.Join(r.Application.Root, r.Handler))
	}, layers.Launch); err != nil {
		return err
	}

	command := fmt.Sprintf(`node %s/server.js`, r.InvokerLayer.Root)

	return r.Layers.WriteMetadata(layers.Metadata{
		Processes: layers.Processes{
			layers.Process{Type: "web", Command: command},
			layers.Process{Type: "function", Command: command},
		},
	})
}

func NewRiffInvoker(build build.Build) (RiffInvoker, bool, error) {
	bp, ok := build.BuildPlan[Dependency]
	if !ok {
		return RiffInvoker{}, false, nil
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return RiffInvoker{}, false, err
	}

	dep, err := deps.Best(Dependency, bp.Version, build.Stack)
	if err != nil {
		return RiffInvoker{}, false, err
	}

	functionJS, ok := bp.Metadata[FunctionArtifact].(string)
	if !ok {
		return RiffInvoker{}, false, fmt.Errorf("node metadata of incorrect type: %v", bp.Metadata[FunctionArtifact])
	}

	return RiffInvoker{
		Application:   build.Application,
		Handler:       functionJS,
		Layers:        build.Layers,
		InvokerLayer:  build.Layers.DependencyLayer(dep),
		FunctionLayer: build.Layers.Layer("function"),
	}, true, nil

}

type marker struct {
	Language string `toml:"language"`
	Function string `toml:"function"`
}

func (m marker) Identity() (string, string) {
	return m.Language, m.Function
}
