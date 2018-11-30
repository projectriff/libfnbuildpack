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

// RiffInvoker represents the Java invoker contributed by the buildpack.
type RiffInvoker struct {
	application application.Application
	handler     string
	layer       layers.DependencyLayer
	layers      layers.Layers
}

// Contribute makes the contribution to the launch layer
func (r RiffInvoker) Contribute() error {
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

// String makes RiffInvoker satisfy the Stringer interface.
func (r RiffInvoker) String() string {
	return fmt.Sprintf("RiffInvoker{ application: %s, handler: %s, layer: %s, layers :%s }",
		r.application, r.handler, r.layer, r.layers)
}

func (r RiffInvoker) command(destination string) string {
	if len(r.handler) > 0 {
		return fmt.Sprintf("java -jar %s $JAVA_OPTS --function.uri='file://%s?handler=%s'",
			destination, r.application.Root, r.handler)
	} else {
		return fmt.Sprintf("java -jar %s $JAVA_OPTS --function.uri='file://%s'",
			destination, r.application.Root)
	}
}

// BuildPlanContribution returns the BuildPlan with requirements for the invoker
func BuildPlanContribution(metadata metadata.Metadata) buildplan.BuildPlan {
	return buildplan.BuildPlan{
		jre.Dependency: buildplan.Dependency{
			Metadata: buildplan.Metadata{jre.LaunchContribution: true},
			Version:  "1.*",
		},
		Dependency: buildplan.Dependency{
			Metadata: buildplan.Metadata{Handler: metadata.Handler,
			},
		},
	}
}

// NewRiffInvoker creates a new RiffInvoker instance. OK is true if build plan contains "riff-invoker-java" dependency,
// otherwise false.
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

	handler, ok := bp.Metadata[Handler].(string)
	if !ok {
		return RiffInvoker{}, false, fmt.Errorf("handler metadata of incorrect type: %v", bp.Metadata[Handler])
	}

	return RiffInvoker{
		build.Application,
		handler,
		build.Layers.DependencyLayer(dep),
		build.Layers,
	}, true, nil
}
