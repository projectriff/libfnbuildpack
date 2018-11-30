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
	"path/filepath"

	"github.com/buildpack/libbuildpack/application"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/build"
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

// RiffCommandInvoker represents the Command invoker contributed by the buildpack.
type RiffCommandInvoker struct {
	// A reference to the user function source tree.
	application application.Application

	// The function executable. Must have exec permissions.
	executable string

	// Provides access to the launch layers, used to craft the process commands.
	layers layers.Layers

	// A dedicated layer for the command invoker itself. Cacheable.
	invokerLayer layers.DependencyLayer

	// A dedicated layer for the function location. Not cacheable, as it changes with the value of executable.
	functionLayer layers.Layer
}

func BuildPlanContribution(metadata metadata.Metadata) buildplan.BuildPlan {
	plans := buildplan.BuildPlan{
		Dependency: buildplan.Dependency{
			Metadata: buildplan.Metadata{
				Command: metadata.Artifact,
			},
		},
	}
	return plans
}

// Contribute makes the contribution to the launch layer
func (r RiffCommandInvoker) Contribute() error {
	if err := r.invokerLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Expanding %s to %s", artifact, layer.Root)
		return helper.ExtractTarGz(artifact, layer.Root, 0)
	}, layers.Cache, layers.Launch); err != nil {
		return err
	}

	if err := r.functionLayer.Contribute(marker{"Command", r.executable}, func(layer layers.Layer) error {
		return layer.OverrideLaunchEnv("FUNCTION_URI", filepath.Join(r.application.Root, r.executable))
	}, layers.Launch); err != nil {
		return err
	}

	command := filepath.Join(r.invokerLayer.Root, functionInvokerExecutable)

	return r.layers.WriteMetadata(layers.Metadata{
		Processes: layers.Processes{
			layers.Process{Type: "web", Command: command},
			layers.Process{Type: "function", Command: command},
		},
	})
}

func NewCommandInvoker(build build.Build) (RiffCommandInvoker, bool, error) {
	bp, ok := build.BuildPlan[Dependency]
	if !ok {
		return RiffCommandInvoker{}, false, nil
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return RiffCommandInvoker{}, false, err
	}

	dep, err := deps.Best(Dependency, bp.Version, build.Stack)
	if err != nil {
		return RiffCommandInvoker{}, false, err
	}

	exec, ok := bp.Metadata[Command].(string)
	if !ok {
		return RiffCommandInvoker{}, false, fmt.Errorf("command metadata of incorrect type: %v", bp.Metadata[Command])
	}

	return RiffCommandInvoker{
		application:   build.Application,
		executable:    exec,
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
