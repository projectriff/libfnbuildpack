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
	"strings"
	"testing"

	"github.com/cloudfoundry/libjavabuildpack"
	"github.com/cloudfoundry/libjavabuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {

	it("fails without metadata", func() {
		f := test.NewEnvironmentFactory(t)
		defer f.Restore()

		f.Console.In(t, "")

		main()

		if *f.ExitStatus != 100 {
			t.Errorf("os.Exit = %d, expected 100", *f.ExitStatus)
		}
	})

	it("passes with metadata and jvm-application", func() {
		f := test.NewEnvironmentFactory(t)
		defer f.Restore()

		f.Console.In(t, "[jvm-application]")

		if err := libjavabuildpack.WriteToFile(strings.NewReader(`handler = "test-handler"`), filepath.Join(f.Application, "riff.toml"), 0644); err != nil {
			t.Fatal(err)
		}

		main()

		if *f.ExitStatus != 0 {
			t.Errorf("os.Exit = %d, expected 0", *f.ExitStatus)
		}
	})

	it("errors with metadata but no application-type", func() {
		f := test.NewEnvironmentFactory(t)
		defer f.Restore()

		f.Console.In(t, "")

		if err := libjavabuildpack.WriteToFile(strings.NewReader(`handler = "test-handler"`), filepath.Join(f.Application, "riff.toml"), 0644); err != nil {
			t.Fatal(err)
		}

		main()

		if *f.ExitStatus != 103 {
			t.Errorf("os.Exit = %d, expected 103", *f.ExitStatus)
		}
	})
}
