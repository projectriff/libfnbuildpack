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
	"path/filepath"
)

const (
	// riffNodeInvokerDependency is a key identifying the node invoker dependency in the build plan.
	riffNodeInvokerDependency = "riff-invoker-node"
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
			Metadata: libbuildpack.BuildPlanDependencyMetadata{"launch": true},
			Version:  "*",
		},
		// Ask for the node invoker
		riffNodeInvokerDependency: libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{
				"fn": metadata.Artifact,
			},
		},
	}
}

// Contribute expands the node invoker tgz and creates launch configurations that run "node server.js"
func (r RiffNodeInvoker) Contribute() error {
	err := r.layer.Contribute(func(artifact string, layer libjavabuildpack.DependencyLaunchLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		return libjavabuildpack.ExtractTarGz(artifact, layer.Root, 0)
	})
	if err != nil {
		return err
	}

	entrypoint := filepath.Join(r.application.Root, r.functionJS)
	command := fmt.Sprintf(`FUNCTION_URI="%s" RIFF_FUNCTION_INVOKER_PROTOCOL=http HOST=0.0.0.0 HTTP_PORT=8080 $NODE_HOME/bin/node %s/server.js`, entrypoint, r.layer.Root)

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

	functionJS, ok := bp.Metadata["fn"].(string)
	if !ok {
		return RiffNodeInvoker{}, false, fmt.Errorf("node metadata of incorrect type: %v", bp.Metadata["fn"])
	}

	return RiffNodeInvoker{
		build.Application,
		functionJS,
		build.Launch,
		build.Launch.DependencyLayer(dep),
	}, true, nil

}
