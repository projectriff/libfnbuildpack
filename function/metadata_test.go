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
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/onsi/gomega"
	"github.com/projectriff/libfnbuildpack/function"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestMetadata(t *testing.T) {
	spec.Run(t, "Metadata", func(t *testing.T, _ spec.G, it spec.S) {

		g := gomega.NewGomegaWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		it("returns false if riff.toml does not exist and RIFF env not set", func() {
			_, ok, err := function.NewMetadata(f.Build.Application, f.Build.Logger)
			g.Expect(ok).To(gomega.BeFalse())
			g.Expect(err).NotTo(gomega.HaveOccurred())
		})

		it("returns metadata if riff.toml exists", func() {
			test.WriteFile(t, filepath.Join(f.Build.Application.Root, "riff.toml"), `
artifact = "toml-artifact"
handler = "toml-handler"
override = "toml-override"
`)

			actual, ok, err := function.NewMetadata(f.Build.Application, f.Build.Logger)
			g.Expect(ok).To(gomega.BeTrue())
			g.Expect(err).NotTo(gomega.HaveOccurred())

			g.Expect(actual).To(gomega.Equal(function.Metadata{
				Artifact: "toml-artifact",
				Handler:  "toml-handler",
				Override: "toml-override",
			}))
		})

		it("returns metadata if RIFF env exists", func() {
			defer test.ReplaceEnv(t, "RIFF", "true")()
			defer test.ReplaceEnv(t, "RIFF_ARTIFACT", "env-artifact")()
			defer test.ReplaceEnv(t, "RIFF_HANDLER", "env-handler")()
			defer test.ReplaceEnv(t, "RIFF_OVERRIDE", "env-override")()

			actual, ok, err := function.NewMetadata(f.Build.Application, f.Build.Logger)
			g.Expect(ok).To(gomega.BeTrue())
			g.Expect(err).NotTo(gomega.HaveOccurred())

			g.Expect(actual).To(gomega.Equal(function.Metadata{
				Artifact: "env-artifact",
				Handler:  "env-handler",
				Override: "env-override",
			}))
		})

		it("environment overrides riff.toml", func() {
			defer test.ReplaceEnv(t, "RIFF", "true")()
			defer test.ReplaceEnv(t, "RIFF_ARTIFACT", "env-artifact")()
			defer test.ReplaceEnv(t, "RIFF_HANDLER", "env-handler")()
			defer test.ReplaceEnv(t, "RIFF_OVERRIDE", "env-override")()
			test.WriteFile(t, filepath.Join(f.Build.Application.Root, "riff.toml"), `
artifact = "toml-artifact"
handler = "toml-handler"
override = "toml-override"
`)

			actual, ok, err := function.NewMetadata(f.Build.Application, f.Build.Logger)
			g.Expect(ok).To(gomega.BeTrue())
			g.Expect(err).NotTo(gomega.HaveOccurred())

			g.Expect(actual).To(gomega.Equal(function.Metadata{
				Artifact: "env-artifact",
				Handler:  "env-handler",
				Override: "env-override",
			}))
		})

	}, spec.Report(report.Terminal{}))
}
