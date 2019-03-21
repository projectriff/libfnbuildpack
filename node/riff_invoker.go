/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/buildpack/libbuildpack"
	"github.com/cloudfoundry/libjavabuildpack"
	nodejs_cnb "github.com/cloudfoundry/nodejs-cnb/build"
	"github.com/projectriff/riff-buildpack"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	// RiffNodeInvokerDependency is a key identifying the node invoker dependency in the build plan.
	RiffNodeInvokerDependency = "riff-invoker-node"

	// functionArtifact is a key identifying the path to the function entrypoint in the build plan.
	FunctionArtifact = "fn"
)

// RiffNodeInvoker represents the Node invoker contributed by the buildpack.
type RiffNodeInvoker struct {
	// A reference to the user function source tree.
	application libbuildpack.Application

	// The file in the function tree that is the entrypoint.
	// May be empty, in which case the function is require()d as a node module.
	functionJS string

	// Provides access to the launch layers, used to craft the process commands.
	launch libjavabuildpack.Launch

	// A dedicated layer for the node invoker itself. Cacheable once npm-installed
	invokerLayer libjavabuildpack.DependencyLaunchLayer

	// A dedicated layer for the function location. Not cacheable, as it changes with the value of functionJS.
	functionLayer libbuildpack.LaunchLayer
}

func BuildPlanContribution(metadata riff_buildpack.Metadata) libbuildpack.BuildPlan {
	return libbuildpack.BuildPlan{
		// Ask the node BP to contribute a node runtime
		nodejs_cnb.NodeDependency: libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{"launch": true, "build": true},
			Version:  "*",
		},
		// Ask for the node invoker
		RiffNodeInvokerDependency: libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{
				FunctionArtifact: metadata.Artifact,
			},
		},
	}
}

// Contribute expands the node invoker tgz and creates launch configurations that run "node server.js"
func (r RiffNodeInvoker) Contribute() error {
	if err := r.invokerLayer.Contribute(func(artifact string, layer libjavabuildpack.DependencyLaunchLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		if e := libjavabuildpack.ExtractTarGz(artifact, layer.Root, 1); e != nil {
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

		if e := layer.WriteProfile("HOST", `export HOST=0.0.0.0`); e != nil {
			return e
		}
		if e := layer.WriteProfile("HTTP_PORT", `export HTTP_PORT=8080`); e != nil {
			return e
		}

		return nil
	}); err != nil {
		return err
	}


	entrypoint := filepath.Join(r.application.Root, r.functionJS)
	if err := r.functionLayer.WriteProfile("FUNCTION_URI", fmt.Sprintf(`export FUNCTION_URI="%s"`, entrypoint)) ; err != nil {
		return err
	}
	if err := r.functionLayer.WriteMetadata(struct {}{}) ; err != nil {
		return err
	}


	command := fmt.Sprintf(`node %s/server.js`, r.invokerLayer.Root)
	return r.launch.WriteMetadata(libbuildpack.LaunchMetadata{
		Processes: libbuildpack.Processes{
			libbuildpack.Process{Type: "web", Command: command}, // TODO: Should be unnecessary once arbitrary process types can be started
			libbuildpack.Process{Type: "function", Command: command},
		},
	})
}

func NewNodeInvoker(build libjavabuildpack.Build) (RiffNodeInvoker, bool, error) {
	bp, ok := build.BuildPlan[RiffNodeInvokerDependency]
	if !ok {
		return RiffNodeInvoker{}, false, nil
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return RiffNodeInvoker{}, false, err
	}

	dep, err := deps.Best(RiffNodeInvokerDependency, bp.Version, build.Stack)
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
		launch:        build.Launch,
		invokerLayer:  build.Launch.DependencyLayer(dep),
		functionLayer: build.Launch.Layer("function"),
	}, true, nil

}
