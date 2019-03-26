/*
 * Copyright 2018 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package function_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff-buildpack/function"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestMetadata(t *testing.T) {
	spec.Run(t, "Metadata", func(t *testing.T, _ spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		it.After(func() {
			os.Unsetenv(function.RiffEnv)
			os.Unsetenv(function.ArtifactEnv)
			os.Unsetenv(function.HandlerEnv)
			os.Unsetenv(function.OverrideEnv)
		})

		it("returns false if riff.toml does not exist and RIFF env not set", func() {
			_, ok, err := function.NewMetadata(f.Build.Application, f.Build.Logger)
			g.Expect(ok).To(BeFalse())
			g.Expect(err).NotTo(HaveOccurred())
		})

		it("returns metadata if riff.toml exists", func() {
			test.WriteFile(t, filepath.Join(f.Build.Application.Root, "riff.toml"), `
artifact = "toml-artifact"
handler = "toml-handler"
override = "toml-override"
`)

			actual, ok, err := function.NewMetadata(f.Build.Application, f.Build.Logger)
			g.Expect(ok).To(BeTrue())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(actual).To(Equal(function.Metadata{
				Artifact: "toml-artifact",
				Handler:  "toml-handler",
				Override: "toml-override",
			}))
		})

		it("returns metadata if RIFF env exists", func() {
			os.Setenv("RIFF", "true")
			os.Setenv("RIFF_ARTIFACT", "env-artifact")
			os.Setenv("RIFF_HANDLER", "env-handler")
			os.Setenv("RIFF_OVERRIDE", "env-override")

			actual, ok, err := function.NewMetadata(f.Build.Application, f.Build.Logger)
			g.Expect(ok).To(BeTrue())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(actual).To(Equal(function.Metadata{
				Artifact: "env-artifact",
				Handler:  "env-handler",
				Override: "env-override",
			}))
		})

		it("environment overrides riff.toml", func() {
			os.Setenv("RIFF", "true")
			os.Setenv("RIFF_ARTIFACT", "env-artifact")
			os.Setenv("RIFF_HANDLER", "env-handler")
			os.Setenv("RIFF_OVERRIDE", "env-override")
			test.WriteFile(t, filepath.Join(f.Build.Application.Root, "riff.toml"), `
artifact = "toml-artifact"
handler = "toml-handler"
override = "toml-override"
`)

			actual, ok, err := function.NewMetadata(f.Build.Application, f.Build.Logger)
			g.Expect(ok).To(BeTrue())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(actual).To(Equal(function.Metadata{
				Artifact: "env-artifact",
				Handler:  "env-handler",
				Override: "env-override",
			}))
		})

	}, spec.Report(report.Terminal{}))
}
