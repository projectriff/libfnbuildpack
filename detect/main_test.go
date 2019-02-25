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
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff-buildpack/function"
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

		it("passes with exactly one riff-invoker detected", func() {
			f.AddBuildPlan("riff-invoker-node", buildplan.Dependency{})
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "riff.toml"), ``)

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))
		})

		it("errors with no riff-invokers detected", func() {
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "riff.toml"), ``)

			code, err := d(f.Detect)
			g.Expect(code).To(Equal(function.Error_DetectedNone))
			g.Expect(err).To(HaveOccurred())
		})

		it("errors with multiple riff-invokers detected", func() {
			f.AddBuildPlan("riff-invoker-node", buildplan.Dependency{})
			f.AddBuildPlan("riff-invoker-java", buildplan.Dependency{})
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "riff.toml"), ``)

			code, err := d(f.Detect)
			g.Expect(code).To(Equal(function.Error_DetectAmbiguity))
			g.Expect(err).To(HaveOccurred())
		})
	}, spec.Report(report.Terminal{}))
}
