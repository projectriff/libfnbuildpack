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
	. "github.com/projectriff/riff-buildpack/plugins"
	"path/filepath"

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

// Contribute makes the contribution to the launch layer
func Contribute(r RiffInvoker) error {
	if err := r.InvokerLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		destination := filepath.Join(layer.Root, layer.ArtifactName())
		layer.Logger.SubsequentLine("Copying to %s", destination)
		return helper.CopyFile(artifact, destination)
	}, layers.Launch); err != nil {
		return err
	}

	command := command(r, filepath.Join(r.InvokerLayer.Root, r.InvokerLayer.ArtifactName()))

	return r.Layers.WriteMetadata(layers.Metadata{
		Processes: layers.Processes{
			layers.Process{Type: "web", Command: command},
			layers.Process{Type: "function", Command: command},
		},
	})
}

func command(r RiffInvoker, destination string) string {
	if len(r.Handler) > 0 {
		return fmt.Sprintf("java -jar %s $JAVA_OPTS --function.uri='file://%s?handler=%s'",
			destination, r.Application.Root, r.Handler)
	} else {
		return fmt.Sprintf("java -jar %s $JAVA_OPTS --function.uri='file://%s'",
			destination, r.Application.Root)
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
		Application:  build.Application,
		Handler:      handler,
		InvokerLayer: build.Layers.DependencyLayer(dep),
		Layers:       build.Layers,
	}, true, nil
}
