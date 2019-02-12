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

package main

import (
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/buildpack/libbuildpack/detect"
	"github.com/cloudfoundry/jvm-application-buildpack/jvmapplication"
	"github.com/cloudfoundry/libcfbuildpack/test"
	nodeCNB "github.com/cloudfoundry/nodejs-cnb/node"
	"github.com/cloudfoundry/npm-cnb/modules"
	"github.com/cloudfoundry/openjdk-buildpack/jre"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff-buildpack/pkg/invoker"
	"github.com/projectriff/riff-buildpack/pkg/invoker/command"
	"github.com/projectriff/riff-buildpack/pkg/invoker/java"
	"github.com/projectriff/riff-buildpack/pkg/invoker/node"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestDetect(t *testing.T) {
	spec.Run(t, "Detect", func(t *testing.T, _ spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.DetectFactory

		it.Before(func() {
			f = test.NewDetectFactory(t)
		})

		it("fails without metadata", func() {
			g.Expect(d(f.Detect)).To(Equal(detect.FailStatusCode))
		})

		it("passes and opts in for the java-invoker if the JVM app BP applied", func() {
			f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "riff.toml"), `handler = "test-handler"`)

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))

			g.Expect(f.Output).To(Equal(buildplan.BuildPlan{
				jre.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{jre.LaunchContribution: true},
				},
				java.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{java.Handler: "test-handler"},
				},
			}))
		})

		it("passes and opts in for the node-invoker if the NPM app BP applied", func() {
			f.AddBuildPlan(modules.Dependency, buildplan.Dependency{})
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "riff.toml"), `artifact = "my.js"`)

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))

			g.Expect(f.Output).To(Equal(buildplan.BuildPlan{
				nodeCNB.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{"launch": true, "build": true},
				},
				node.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{node.FunctionArtifact: "my.js"},
				},
			}))
		})

		it("passes and opts in for the node-invoker if the NPM app BP did not apply, but artifact is .js", func() {
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "my.js"), "module.exports = x => x**2")
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "riff.toml"), `artifact = "my.js"`)

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))

			g.Expect(f.Output).To(Equal(buildplan.BuildPlan{
				nodeCNB.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{"launch": true, "build": true},
				},
				node.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{node.FunctionArtifact: "my.js"},
				},
			}))
		})

		it("passes and opts in for the command-invoker if the artifact is executable", func() {
			test.WriteFileWithPerm(t, filepath.Join(f.Detect.Application.Root, "fn.sh"), 0755 /*<-executable*/, "some bash")
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "riff.toml"), `artifact = "fn.sh"`)

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))

			g.Expect(f.Output).To(Equal(buildplan.BuildPlan{
				command.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{command.Command: "fn.sh"},
				},
			}))
		})

		it("fails if ambiguity", func() {
			f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
			f.AddBuildPlan(modules.Dependency, buildplan.Dependency{})
			test.WriteFileWithPerm(t, filepath.Join(f.Detect.Application.Root, "fn.sh"), 0755 /*<-executable*/, "some bash")
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "riff.toml"), `artifact = "fn.sh"`)

			code, err := d(f.Detect)
			g.Expect(code).To(Equal(invoker.Error_DetectAmbiguity))
			g.Expect(err).To(HaveOccurred())
		})

		it("override resolves ambiguity", func() {
			f.AddBuildPlan(jvmapplication.Dependency, buildplan.Dependency{})
			f.AddBuildPlan(modules.Dependency, buildplan.Dependency{})
			test.WriteFileWithPerm(t, filepath.Join(f.Detect.Application.Root, "fn.sh"), 0755 /*<-executable*/, "some bash")
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "riff.toml"), `artifact = "fn.sh"
override = "java"`)

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))

			g.Expect(f.Output).To(Equal(buildplan.BuildPlan{
				jre.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{jre.LaunchContribution: true},
				},
				java.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{java.Handler: ""},
				},
			}))
		})

		it("errors with metadata but no application-type", func() {
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "riff.toml"), `handler = "test-handler"`)

			code, err := d(f.Detect)
			g.Expect(code).To(Equal(invoker.Error_DetectedNone))
			g.Expect(err).To(HaveOccurred())
		})
	}, spec.Report(report.Terminal{}))
}
