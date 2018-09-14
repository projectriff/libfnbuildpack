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
	"reflect"
	"testing"

	"github.com/buildpack/libbuildpack"
	"github.com/projectriff/riff-buildpack"
	"github.com/projectriff/riff-buildpack/java"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestRiffInvoker(t *testing.T) {
	spec.Run(t, "RiffInvoker", testRiffInvoker, spec.Report(report.Terminal{}))
}

func testRiffInvoker(t *testing.T, when spec.G, it spec.S) {

	when("returning BuildPlan", func() {

		it("contains openjdk-jre", func() {
			bp := java.RiffInvoker{}.BuildPlanContribution(riff_buildpack.Metadata{})

			// TODO use constants for openjdk-jre and launch
			actual := bp["openjdk-jre"]

			expected := libbuildpack.BuildPlanDependency{
				Metadata: libbuildpack.BuildPlanDependencyMetadata{"launch": true},
				Version:  "1.*",
			}

			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("BuildPlan[\"openjdk-jre\"] = %s, expected = %s", actual, expected)
			}
		})

		it("contains riff-invoker-java", func() {
			bp := java.RiffInvoker{}.BuildPlanContribution(riff_buildpack.Metadata{Handler: "test-handler"})

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
	})
}
