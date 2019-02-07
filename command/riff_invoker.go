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

package command

import (
	"fmt"
	. "github.com/projectriff/riff-buildpack/plugins"
	"path/filepath"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/projectriff/riff-buildpack/metadata"
)

const (
	// Dependency is a key identifying the command invoker dependency in the build plan.
	Dependency = "riff-invoker-command"

	// command is the key identifying the command executable in the build plan.
	Command = "command"

	// functionInvokerExecutable is the name of the function invoker in the tgz dependency
	functionInvokerExecutable = "command-function-invoker"
)

func BuildPlanContribution(detect detect.Detect, metadata metadata.Metadata) buildplan.BuildPlan {
	r := detect.BuildPlan[Dependency]
	if r.Metadata == nil {
		r.Metadata = buildplan.Metadata{}
	}
	r.Metadata[Command] = metadata.Artifact

	return buildplan.BuildPlan{Dependency: r}
}

// Contribute makes the contribution to the launch layer
func Contribute(r RiffInvoker) error {
	if err := r.InvokerLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Expanding %s to %s", artifact, layer.Root)
		return helper.ExtractTarGz(artifact, layer.Root, 0)
	}, layers.Cache, layers.Launch); err != nil {
		return err
	}

	if err := r.FunctionLayer.Contribute(marker{"Command", r.Handler}, func(layer layers.Layer) error {
		return layer.OverrideLaunchEnv("FUNCTION_URI", filepath.Join(r.Application.Root, r.Handler))
	}, layers.Launch); err != nil {
		return err
	}

	command := filepath.Join(r.InvokerLayer.Root, functionInvokerExecutable)

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

	exec, ok := bp.Metadata[Command].(string)
	if !ok {
		return RiffInvoker{}, false, fmt.Errorf("command metadata of incorrect type: %v", bp.Metadata[Command])
	}

	return RiffInvoker{
		Application:   build.Application,
		Handler:       exec,
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
