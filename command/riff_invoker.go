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
	"github.com/buildpack/libbuildpack"
	"github.com/cloudfoundry/libjavabuildpack"
	"github.com/projectriff/riff-buildpack"
	"path/filepath"
)

const (
	// RiffCommandInvokerDependency is a key identifying the command invoker dependency in the build plan.
	RiffCommandInvokerDependency = "riff-invoker-command"

	// command is the key identifying the command executable in the build plan.
	Command = "command"

	// functionInvokerExecutable is the name of the function invoker in the tgz dependency
	functionInvokerExecutable = "command-function-invoker"
)

// RiffCommandInvoker represents the Command invoker contributed by the buildpack.
type RiffCommandInvoker struct {
	// A reference to the user function source tree.
	application libbuildpack.Application

	// The function executable. Must have exec permissions.
	executable string

	// Provides access to the launch layers, used to craft the process commands.
	launch libjavabuildpack.Launch

	// A dedicated layer for the command invoker itself. Cacheable.
	invokerLayer libjavabuildpack.DependencyLaunchLayer

	// A dedicated layer for the function location. Not cacheable, as it changes with the value of executable.
	functionLayer libbuildpack.LaunchLayer
}

func BuildPlanContribution(metadata riff_buildpack.Metadata) libbuildpack.BuildPlan {
	plans := libbuildpack.BuildPlan{
		RiffCommandInvokerDependency: libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{
				Command: metadata.Artifact,
			},
		},
	}
	return plans
}

// Contribute makes the contribution to the launch layer
func (r RiffCommandInvoker) Contribute() error {
	if err := r.invokerLayer.Contribute(func(artifact string, layer libjavabuildpack.DependencyLaunchLayer) error {
		layer.Logger.SubsequentLine("Expanding %s to %s", artifact, layer.Root)
		return libjavabuildpack.ExtractTarGz(artifact, layer.Root, 0)
	}); err != nil {
		return err
	}

	functionURI := filepath.Join(r.application.Root, r.executable)
	if err := r.functionLayer.WriteProfile("FUNCTION_URI", fmt.Sprintf(`export FUNCTION_URI='%s'`, functionURI)) ; err != nil {
		return err
	}
	if err := r.functionLayer.WriteMetadata(struct{}{}) ; err != nil {
		return err
	}

	command := filepath.Join(r.invokerLayer.Root, functionInvokerExecutable)
	return r.launch.WriteMetadata(libbuildpack.LaunchMetadata{
		Processes: libbuildpack.Processes{
			libbuildpack.Process{Type: "web", Command: command},
			libbuildpack.Process{Type: "function", Command: command},
		},
	})
}

func NewCommandInvoker(build libjavabuildpack.Build) (RiffCommandInvoker, bool, error) {
	bp, ok := build.BuildPlan[RiffCommandInvokerDependency]
	if !ok {
		return RiffCommandInvoker{}, false, nil
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return RiffCommandInvoker{}, false, err
	}

	dep, err := deps.Best(RiffCommandInvokerDependency, bp.Version, build.Stack)
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
		launch:        build.Launch,
		invokerLayer:  build.Launch.DependencyLayer(dep),
		functionLayer: build.Launch.Layer("function"),
	}, true, nil

}
