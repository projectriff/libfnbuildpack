/*
 * Copyright 2018-2020 the original author or authors.
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

package libfnbuildpack_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libpak"
	"github.com/projectriff/libfnbuildpack"
	"github.com/sclevine/spec"
)

func testMetadata(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
	)

	it.Before(func() {
		var err error

		path, err = ioutil.TempDir("", "metadata")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("IsRiff", func() {

		context(`RIFF=""`, func() {

			it.Before(func() {
				Expect(os.Setenv("RIFF", "")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("RIFF")).To(Succeed())
			})

			it("returns false if environment variable is empty", func() {
				Expect(libfnbuildpack.IsRiff(path, libpak.ConfigurationResolver{})).To(BeFalse())
			})

		})

		context(`RIFF="true"`, func() {

			it.Before(func() {
				Expect(os.Setenv("RIFF", "true")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("RIFF")).To(Succeed())
			})

			it("returns true if environment variable is set", func() {
				Expect(libfnbuildpack.IsRiff(path, libpak.ConfigurationResolver{})).To(BeTrue())
			})

		})

		it("returns true if riff.toml exists", func() {
			Expect(ioutil.WriteFile(filepath.Join(path, "riff.toml"), []byte{}, 0644)).To(Succeed())

			Expect(libfnbuildpack.IsRiff(path, libpak.ConfigurationResolver{})).To(BeTrue())
		})

		it("return false if no indication", func() {
			Expect(libfnbuildpack.IsRiff(path, libpak.ConfigurationResolver{})).To(BeFalse())
		})

	})

	context("Metadata", func() {

		it("returns empty if riff.toml doesn't exist", func() {
			Expect(libfnbuildpack.Metadata(path, libpak.ConfigurationResolver{})).To(BeEmpty())
		})

		it("returns contents of riff.toml", func() {
			Expect(ioutil.WriteFile(filepath.Join(path, "riff.toml"), []byte(`
artifact = "test-artifact-1"
handler  = "test-handler-1"
`), 0644)).To(Succeed())

			Expect(libfnbuildpack.Metadata(path, libpak.ConfigurationResolver{})).To(Equal(map[string]interface{}{
				"artifact": "test-artifact-1",
				"handler":  "test-handler-1",
			}))
		})

		context(`$RIFF_ARTIFACT=""`, func() {

			it.Before(func() {
				Expect(os.Setenv("RIFF_ARTIFACT", "")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("RIFF_ARTIFACT")).To(Succeed())
			})

			it("return contents of riff.toml", func() {
				Expect(ioutil.WriteFile(filepath.Join(path, "riff.toml"), []byte(`
artifact = "test-artifact-1"
handler  = "test-handler-1"
`), 0644)).To(Succeed())

				Expect(libfnbuildpack.Metadata(path, libpak.ConfigurationResolver{})).To(Equal(map[string]interface{}{
					"artifact": "test-artifact-1",
					"handler":  "test-handler-1",
				}))
			})

		})


		context(`$RIFF_ARTIFACT="test-artifact-2"`, func() {

			it.Before(func() {
				Expect(os.Setenv("RIFF_ARTIFACT", "test-artifact-2")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("RIFF_ARTIFACT")).To(Succeed())
			})

			it("return contents of $RIFF_ARTIFACT", func() {
				Expect(ioutil.WriteFile(filepath.Join(path, "riff.toml"), []byte(`
artifact = "test-artifact-1"
handler  = "test-handler-1"
`), 0644)).To(Succeed())

				Expect(libfnbuildpack.Metadata(path, libpak.ConfigurationResolver{})).To(Equal(map[string]interface{}{
					"artifact": "test-artifact-2",
					"handler":  "test-handler-1",
				}))
			})

		})

		context(`$RIFF_HANDLER="test-handler-2"`, func() {

			it.Before(func() {
				Expect(os.Setenv("RIFF_HANDLER", "")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("RIFF_HANDLER")).To(Succeed())
			})

			it("return contents of $RIFF_HANDLER", func() {
				Expect(ioutil.WriteFile(filepath.Join(path, "riff.toml"), []byte(`
artifact = "test-artifact-1"
handler  = "test-handler-1"
`), 0644)).To(Succeed())

				Expect(libfnbuildpack.Metadata(path, libpak.ConfigurationResolver{})).To(Equal(map[string]interface{}{
					"artifact": "test-artifact-1",
					"handler":  "test-handler-1",
				}))
			})

		})

		context(`$RIFF_HANDLER="test-handler-2"`, func() {

			it.Before(func() {
				Expect(os.Setenv("RIFF_HANDLER", "test-handler-2")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("RIFF_HANDLER")).To(Succeed())
			})

			it("return contents of $RIFF_HANDLER", func() {
				Expect(ioutil.WriteFile(filepath.Join(path, "riff.toml"), []byte(`
artifact = "test-artifact-1"
handler  = "test-handler-1"
`), 0644)).To(Succeed())

				Expect(libfnbuildpack.Metadata(path, libpak.ConfigurationResolver{})).To(Equal(map[string]interface{}{
					"artifact": "test-artifact-1",
					"handler":  "test-handler-2",
				}))
			})

		})

	})

}
