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

package java_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/buildpack/libbuildpack"
	"github.com/cloudfoundry/libjavabuildpack/test"
	"github.com/cloudfoundry/openjdk-buildpack"
	"github.com/projectriff/riff-buildpack"
	"github.com/projectriff/riff-buildpack/java"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestRiffInvoker(t *testing.T) {
	spec.Run(t, "RiffInvoker", testRiffInvoker, spec.Report(report.Terminal{}))
}

func testRiffInvoker(t *testing.T, when spec.G, it spec.S) {

	it("contains openjdk-jre", func() {
		bp := java.BuildPlanContribution(riff_buildpack.Metadata{})

		actual := bp[openjdk_buildpack.JREDependency]

		expected := libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{openjdk_buildpack.LaunchContribution: true},
			Version:  "1.*",
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("BuildPlan[\"openjdk-jre\"] = %s, expected = %s", actual, expected)
		}
	})

	it("contains riff-invoker-java", func() {
		bp := java.BuildPlanContribution(riff_buildpack.Metadata{Handler: "test-handler"})

		actual := bp[java.RiffInvokerDependency]

		expected := libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{
				java.Handler: "test-handler",
			},
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("BuildPlan[\"riff-invoker-java\"] = %s, expected = %s", actual, expected)
		}
	})

	it("returns true if build plan exists", func() {
		f := test.NewBuildFactory(t)
		f.AddDependency(t, java.RiffInvokerDependency, "stub-invoker.jar")
		f.AddBuildPlan(t, java.RiffInvokerDependency, libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{java.Handler: "test-handler"},
		})

		_, ok, err := java.NewRiffInvoker(f.Build)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Errorf("NewRiffInvoker = %t, expected true", ok)
		}
	})

	it("returns false if build plan does not exist", func() {
		f := test.NewBuildFactory(t)

		_, ok, err := java.NewRiffInvoker(f.Build)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Errorf("NewRiffInvoker = %t, expected false", ok)
		}
	})

	it("contributes invoker", func() {
		f := test.NewBuildFactory(t)
		f.AddDependency(t, java.RiffInvokerDependency, "stub-invoker.jar")
		f.AddBuildPlan(t, java.RiffInvokerDependency, libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{java.Handler: "test-handler"},
		})

		r, _, err := java.NewRiffInvoker(f.Build)
		if err != nil {
			t.Fatal(err)
		}

		if err := r.Contribute(); err != nil {
			t.Fatal(err)
		}

		layerRoot := filepath.Join(f.Build.Launch.Root, "riff-invoker-java")
		test.BeFileLike(t, filepath.Join(layerRoot, "stub-invoker.jar"), 0644, "")

		var actual libbuildpack.LaunchMetadata
		_, err = toml.DecodeFile(filepath.Join(f.Build.Launch.Root, "launch.toml"), &actual)
		if err != nil {
			t.Fatal(err)
		}

		command := fmt.Sprintf("java -jar %s $JAVA_OPTS --function.uri='file://%s?handler=test-handler'",
			filepath.Join(layerRoot, "stub-invoker.jar"), f.Build.Application.Root)

		expected := libbuildpack.LaunchMetadata{
			Processes: libbuildpack.Processes{
				libbuildpack.Process{Type: "web", Command: command},
				libbuildpack.Process{Type: "function", Command: command},
			},
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("launch.toml = %s, expected %s", actual, expected)
		}
	})
}
