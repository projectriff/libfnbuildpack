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

package riff_buildpack_test

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/cloudfoundry/libjavabuildpack"
	"github.com/cloudfoundry/libjavabuildpack/test"
	"github.com/projectriff/riff-buildpack"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestMetadata(t *testing.T) {
	spec.Run(t, "Metadata", testMetadata, spec.Report(report.Terminal{}))
}

func testMetadata(t *testing.T, when spec.G, it spec.S) {

	it("returns false if riff.toml does not exist", func() {
		f := test.NewBuildFactory(t)

		_, ok, err := riff_buildpack.NewMetadata(f.Build.Application, f.Build.Logger)
		if err != nil {
			t.Fatal(err)
		}

		if ok {
			t.Errorf("NewMetadata = %t, expected false", ok)
		}
	})

	it("returns metadata if riff.toml does exist", func() {
		f := test.NewBuildFactory(t)

		metadata := filepath.Join(f.Build.Application.Root, "riff.toml")
		libjavabuildpack.WriteToFile(strings.NewReader(`handler = "test-handler"`), metadata, 0644)

		actual, ok, err := riff_buildpack.NewMetadata(f.Build.Application, f.Build.Logger)
		if err != nil {
			t.Fatal(err)
		}

		if !ok {
			t.Errorf("NewMetadata = %t, expected true", ok)
		}

		expected := riff_buildpack.Metadata{Handler: "test-handler"}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("NewMetadata = %s, expected %s", actual, expected)
		}
	})
}
