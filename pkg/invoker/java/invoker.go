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

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/projectriff/riff-buildpack/pkg/invoker"
)

const (
	// Dependency is the key identifying the riff java invoker in the buildpack plan.
	Dependency = "riff-invoker-java"
	// Handler is the key identifying the riff handler metadata in the build plan
	Handler = "handler"
	// invokerMainClass is the class name to run
	invokerMainClass = "org.springframework.boot.loader.JarLauncher"
)

// RiffJavaInvoker represents the Java invoker contributed by the buildpack.
type RiffJavaInvoker struct {
	// A reference to the user function.
	application application.Application

	// Optional reference to the java class implementing the function.
	handler string

	// Provides access to the launch layers, used to craft the process commands.
	layers layers.Layers

	// A dedicated layer for the java invoker. Cacheable once unzipped.
	invokerLayer layers.DependencyLayer

	// A dedicated layer for the function location. Not cacheable as it changes with the value of handler.
	functionLayer layers.Layer
}

// Contribute makes the contribution to the launch layer
func (r RiffJavaInvoker) Contribute() error {
	if err := r.invokerLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Unzipping java invoker to %s", layer.Root)
		return helper.ExtractZip(artifact, layer.Root, 0)
	}, layers.Launch); err != nil {
		return err
	}

	if err := r.functionLayer.Contribute(marker{"Java", r.handler}, func(layer layers.Layer) error {
		if len(r.handler) > 0 {
			return layer.OverrideLaunchEnv("FUNCTION_URI", fmt.Sprintf("file://%s?handler=%s", r.application.Root, r.handler))
		} else {
			return layer.OverrideLaunchEnv("FUNCTION_URI", fmt.Sprintf("file://%s", r.application.Root))
		}
	}, layers.Launch); err != nil {
		return err
	}

	command := fmt.Sprintf("java -cp %s $JAVA_OPTS %s", r.invokerLayer.Root, invokerMainClass)

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
		r.application, r.handler, r.invokerLayer, r.layers)
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

// NewJavaInvoker creates a new RiffJavaInvoker instance. OK is true if build plan contains "riff-invoker-java" dependency,
// otherwise false.
func NewJavaInvoker(build build.Build) (invoker.RiffInvoker, bool, error) {
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
		application:   build.Application,
		handler:       handler,
		layers:        build.Layers,
		invokerLayer:  build.Layers.DependencyLayer(dep),
		functionLayer: build.Layers.Layer("function"),
	}, true, nil
}

type marker struct {
	Language string `toml:"language"`
	Handler  string `toml:"handler"`
}

func (m marker) Identity() (string, string) {
	return m.Language, m.Handler
}
