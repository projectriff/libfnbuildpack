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

package metadata_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff-buildpack/pkg/metadata"
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

		it("returns false if riff.toml does not exist", func() {
			_, ok, err := metadata.NewMetadata(f.Build.Application, f.Build.Logger)
			g.Expect(ok).To(BeFalse())
			g.Expect(err).NotTo(HaveOccurred())
		})

		it("returns metadata if riff.toml does exist", func() {
			test.WriteFile(t, filepath.Join(f.Build.Application.Root, "riff.toml"), `handler = "test-handler"`)

			actual, ok, err := metadata.NewMetadata(f.Build.Application, f.Build.Logger)
			g.Expect(ok).To(BeTrue())
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(actual).To(Equal(metadata.Metadata{Handler: "test-handler"}))
		})
	}, spec.Report(report.Terminal{}))
}
