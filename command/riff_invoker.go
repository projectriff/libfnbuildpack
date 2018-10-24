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
	"os"
	"path/filepath"
)

const (
	// RiffCommandInvokerDependency is a key identifying the command invoker dependency in the build plan.
	riffCommandInvokerDependency = "riff-invoker-command"

	// command is the key identifying the command executable in the build plan.
	command = "command"
)

// RiffCommandInvoker represents the Command invoker contributed by the buildpack.
type RiffCommandInvoker struct {
	application libbuildpack.Application
	executable  string
	launch      libjavabuildpack.Launch
	layer       libjavabuildpack.DependencyLaunchLayer
}

func BuildPlanContribution(metadata riff_buildpack.Metadata) libbuildpack.BuildPlan {
	plans := libbuildpack.BuildPlan{
		riffCommandInvokerDependency: libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{
				command: metadata.Artifact,
			},
		},
	}
	return plans
}

// Contribute makes the contribution to the launch layer
func (r RiffCommandInvoker) Contribute() error {
	err := r.layer.Contribute(func(artifact string, layer libjavabuildpack.DependencyLaunchLayer) error {
		destination := filepath.Join(layer.Root, layer.ArtifactName())
		layer.Logger.SubsequentLine("Copying to %s", destination)

		if e := libjavabuildpack.CopyFile(artifact, destination); e != nil {
			return e
		}
		return os.Chmod(destination, 0755)
	})
	if err != nil {
		return err
	}

	command := r.command(filepath.Join(r.layer.Root, r.layer.ArtifactName()))

	return r.launch.WriteMetadata(libbuildpack.LaunchMetadata{
		Processes: libbuildpack.Processes{
			libbuildpack.Process{Type: "web", Command: command}, // TODO: Should be unnecessary once arbitrary process types can be started
			libbuildpack.Process{Type: "function", Command: command},
		},
	})
}

func (r RiffCommandInvoker) command(invokerPath string) string {
	fn := filepath.Join(r.application.Root, r.executable)
	return fmt.Sprintf(`FUNCTION_URI=%s %s`, fn, invokerPath)
}

func NewCommandInvoker(build libjavabuildpack.Build) (RiffCommandInvoker, bool, error) {
	bp, ok := build.BuildPlan[riffCommandInvokerDependency]
	if !ok {
		return RiffCommandInvoker{}, false, nil
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return RiffCommandInvoker{}, false, err
	}

	dep, err := deps.Best(riffCommandInvokerDependency, bp.Version, build.Stack)
	if err != nil {
		return RiffCommandInvoker{}, false, err
	}

	exec, ok := bp.Metadata[command].(string)
	if !ok {
		return RiffCommandInvoker{}, false, fmt.Errorf("command metadata of incorrect type: %v", bp.Metadata[command])
	}

	return RiffCommandInvoker{
		build.Application,
		exec,
		build.Launch,
		build.Launch.DependencyLayer(dep),
	}, true, nil

}
