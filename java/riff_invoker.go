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

	"github.com/buildpack/libbuildpack/application"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/openjdk-buildpack/jre"
	"github.com/projectriff/riff-buildpack/metadata"
)

const (
	// Dependency is the key identifying the riff java invoker in the buildpack plan.
	Dependency = "riff-invoker-java"
	// Handler is the key identifying the riff handler metadata in the build plan
	Handler = "handler"
)

// RiffJavaInvoker represents the Java invoker contributed by the buildpack.
type RiffJavaInvoker struct {
	application application.Application
	handler     string
	layer       layers.DependencyLayer
	layers      layers.Layers
}

// Contribute makes the contribution to the launch layer
func (r RiffJavaInvoker) Contribute() error {
	if err := r.layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		destination := filepath.Join(layer.Root, layer.ArtifactName())
		layer.Logger.SubsequentLine("Copying to %s", destination)
		return helper.CopyFile(artifact, destination)
	}, layers.Launch); err != nil {
		return err
	}

	command := r.command(filepath.Join(r.layer.Root, r.layer.ArtifactName()))

	return r.layers.WriteMetadata(layers.Metadata{
		Processes: layers.Processes{
			layers.Process{Type: "web", Command: command},
			layers.Process{Type: "function", Command: command},
		},
	})
}

// String makes RiffJavaInvoker satisfy the Stringer interface.
func (r RiffJavaInvoker) String() string {
	return fmt.Sprintf("RiffJavaInvoker{ application: %s, handler: %s, layer: %s, layers :%s }",
		r.application, r.handler, r.layer, r.layers)
}

func (r RiffJavaInvoker) command(destination string) string {
	if len(r.handler) > 0 {
		return fmt.Sprintf("java -jar %s $JAVA_OPTS --function.uri='file://%s?handler=%s'",
			destination, r.application.Root, r.handler)
	} else {
		return fmt.Sprintf("java -jar %s $JAVA_OPTS --function.uri='file://%s'",
			destination, r.application.Root)
	}
}

// BuildPlanContribution returns the BuildPlan with requirements for the invoker
func BuildPlanContribution(detect detect.Detect, metadata metadata.Metadata) buildplan.BuildPlan {
	j := detect.BuildPlan[jre.Dependency]
	if j.Metadata == nil {
		j.Metadata = buildplan.Metadata{}
	}
	j.Metadata[jre.LaunchContribution] = true

	r := detect.BuildPlan[Dependency]
	if r.Metadata == nil {
		r.Metadata = buildplan.Metadata{}
	}
	r.Metadata[Handler] = metadata.Handler

	return buildplan.BuildPlan{jre.Dependency: j, Dependency: r}
}

// NewJavaInvoker creates a new RiffJavaInvoker instance. OK is true if build plan contains "riff-invoker-java" dependency,
// otherwise false.
func NewJavaInvoker(build build.Build) (RiffJavaInvoker, bool, error) {
	bp, ok := build.BuildPlan[Dependency]
	if !ok {
		return RiffJavaInvoker{}, false, nil
	}

	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return RiffJavaInvoker{}, false, err
	}

	dep, err := deps.Best(Dependency, bp.Version, build.Stack)
	if err != nil {
		return RiffJavaInvoker{}, false, err
	}

	handler, ok := bp.Metadata[Handler].(string)
	if !ok {
		return RiffJavaInvoker{}, false, fmt.Errorf("handler metadata of incorrect type: %v", bp.Metadata[Handler])
	}

	return RiffJavaInvoker{
		build.Application,
		handler,
		build.Layers.DependencyLayer(dep),
		build.Layers,
	}, true, nil
}
