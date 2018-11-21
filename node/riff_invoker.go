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
	"github.com/buildpack/libbuildpack"
	"github.com/cloudfoundry/libjavabuildpack"
	nodejs_cnb "github.com/cloudfoundry/nodejs-cnb/build"
	"github.com/projectriff/riff-buildpack"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	// riffNodeInvokerDependency is a key identifying the node invoker dependency in the build plan.
	riffNodeInvokerDependency = "riff-invoker-node"

	// functionArtifact is a key identifying the path to the function entrypoint in the build plan.
	functionArtifact = "fn"
)

// RiffNodeInvoker represents the Node invoker contributed by the buildpack.
type RiffNodeInvoker struct {
	application libbuildpack.Application
	functionJS  string
	launch      libjavabuildpack.Launch
	layer       libjavabuildpack.DependencyLaunchLayer
}

func BuildPlanContribution(metadata riff_buildpack.Metadata) libbuildpack.BuildPlan {
	return libbuildpack.BuildPlan{
		// Ask the node BP to contribute a node runtime
		nodejs_cnb.NodeDependency: libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{"launch": true, "build": true},
			Version:  "*",
		},
		// Ask for the node invoker
		riffNodeInvokerDependency: libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{
				functionArtifact: metadata.Artifact,
			},
		},
	}
}

// Contribute expands the node invoker tgz and creates launch configurations that run "node server.js"
func (r RiffNodeInvoker) Contribute() error {
	err := r.layer.Contribute(func(artifact string, layer libjavabuildpack.DependencyLaunchLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		if e := libjavabuildpack.ExtractTarGz(artifact, layer.Root, 1) ; e != nil {
			return e
		}
		layer.Logger.SubsequentLine("npm-installing the node invoker")
		cmd := exec.Command("npm", "install", "--production")
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.Dir = layer.Root
		if e := cmd.Run() ; e != nil {
			return e
		}

		if e := layer.WriteProfile("host", `export HOST=0.0.0.0`) ; e != nil {
			return e
		}
		if e := layer.WriteProfile("http-port", `export HTTP_PORT=8080`) ; e != nil {
			return e
		}

		return nil
	})
	if err != nil {
		return err
	}

	entrypoint := filepath.Join(r.application.Root, r.functionJS)
	command := fmt.Sprintf(`FUNCTION_URI="%s" node %s/server.js`, entrypoint, r.layer.Root)

	return r.launch.WriteMetadata(libbuildpack.LaunchMetadata{
		Processes: libbuildpack.Processes{
			libbuildpack.Process{Type: "web", Command: command}, // TODO: Should be unnecessary once arbitrary process types can be started
			libbuildpack.Process{Type: "function", Command: command},
		},
	})
}

func NewNodeInvoker(build libjavabuildpack.Build) (RiffNodeInvoker, bool, error) {
	bp, ok := build.BuildPlan[riffNodeInvokerDependency]
	if !ok {
		return RiffNodeInvoker{}, false, nil
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return RiffNodeInvoker{}, false, err
	}

	dep, err := deps.Best(riffNodeInvokerDependency, bp.Version, build.Stack)
	if err != nil {
		return RiffNodeInvoker{}, false, err
	}

	functionJS, ok := bp.Metadata[functionArtifact].(string)
	if !ok {
		return RiffNodeInvoker{}, false, fmt.Errorf("node metadata of incorrect type: %v", bp.Metadata[functionArtifact])
	}

	return RiffNodeInvoker{
		build.Application,
		functionJS,
		build.Launch,
		build.Launch.DependencyLayer(dep),
	}, true, nil

}
