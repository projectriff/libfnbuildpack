/*
 * Copyright 2018 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package java

import (
	"fmt"
	"path/filepath"

	"github.com/buildpack/libbuildpack"
	"github.com/cloudfoundry/libjavabuildpack"
	"github.com/cloudfoundry/openjdk-buildpack"
	"github.com/projectriff/riff-buildpack"
)

const (
	// Handler is the key identifying the riff handler metadata in the build plan
	Handler = "handler"

	// RiffInvokerDependency is the key identifying the riff java invoker in the buildpack plan.
	RiffInvokerDependency = "riff-invoker-java"
)

// RiffInvoker represents the Java invoker contributed by the buildpack.
type RiffInvoker struct {
	application libbuildpack.Application
	handler     string
	launch      libjavabuildpack.Launch
	layer       libjavabuildpack.DependencyLaunchLayer
}

// Contribute makes the contribution to the launch layer
func (r RiffInvoker) Contribute() error {
	err := r.layer.Contribute(func(artifact string, layer libjavabuildpack.DependencyLaunchLayer) error {
		destination := filepath.Join(layer.Root, layer.ArtifactName())
		layer.Logger.SubsequentLine("Copying to %s", destination)
		return libjavabuildpack.CopyFile(artifact, destination)
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

// String makes RiffInvoker satisfy the Stringer interface.
func (r RiffInvoker) String() string {
	return fmt.Sprintf("RiffInvoker{ application: %s, handler: %s, launch: %s, layer :%s }",
		r.application, r.handler, r.launch, r.layer)
}

func (r RiffInvoker) command(destination string) string {
	return fmt.Sprintf("java -jar %s $JAVA_OPTS --function.uri='file://%s?handler=%s'",
		destination, r.application.Root, r.handler)
}

// BuildPlanContribution returns the BuildPlan with requirements for the invoker
func BuildPlanContribution(metadata riff_buildpack.Metadata) libbuildpack.BuildPlan {
	return libbuildpack.BuildPlan{
		openjdk_buildpack.JREDependency: libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{openjdk_buildpack.LaunchContribution: true},
			Version:  "1.*",
		},
		RiffInvokerDependency: libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{
				Handler: metadata.Handler,
			},
		},
	}
}

// NewRiffInvoker creates a new RiffInvoker instance. OK is true if build plan contains "riff-invoker-java" dependency,
// otherwise false.
func NewRiffInvoker(build libjavabuildpack.Build) (RiffInvoker, bool, error) {
	bp, ok := build.BuildPlan[RiffInvokerDependency]
	if !ok {
		return RiffInvoker{}, false, nil
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return RiffInvoker{}, false, err
	}

	dep, err := deps.Best(RiffInvokerDependency, bp.Version, build.Stack)
	if err != nil {
		return RiffInvoker{}, false, err
	}

	handler, ok := bp.Metadata[Handler].(string)
	if !ok {
		return RiffInvoker{}, false, fmt.Errorf("handler metadata of incorrect type: %v", bp.Metadata[Handler])
	}

	return RiffInvoker{
		build.Application,
		handler,
		build.Launch,
		build.Launch.DependencyLayer(dep),
	}, true, nil
}
