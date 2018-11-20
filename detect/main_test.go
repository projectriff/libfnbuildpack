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
	"fmt"
	"github.com/buildpack/libbuildpack"
	npmdetect "github.com/cloudfoundry/npm-cnb/detect"
	"github.com/projectriff/riff-buildpack/command"
	"github.com/projectriff/riff-buildpack/java"
	"github.com/projectriff/riff-buildpack/node"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudfoundry/jvm-application-buildpack"
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

	it("passes and opts in for the java-invoker if the JVM app BP applied", func() {
		f := test.NewEnvironmentFactory(t)
		defer f.Restore()

		f.Console.In(t, fmt.Sprintf("[%s]", jvm_application_buildpack.JVMApplication))

		if err := libjavabuildpack.WriteToFile(strings.NewReader(`handler = "test-handler"`), filepath.Join(f.Application, "riff.toml"), 0644); err != nil {
			t.Fatal(err)
		}

		main()

		if *f.ExitStatus != 0 {
			t.Errorf("os.Exit = %d, expected 0", *f.ExitStatus)
		}

		if bp, err := libbuildpack.NewBuildPlan(strings.NewReader(f.Console.Out(t)), libbuildpack.NewLogger(nil, nil)) ; err != nil {
			t.Fatal(err)
		} else {
			if j, ok := bp[java.RiffInvokerDependency] ; !ok {
				t.Error("expected the java invoker to be added: ", bp)
			} else {
				if j.Metadata[java.Handler] != "test-handler" {
					t.Error("handler key not set: ", bp)
				}
			}
		}
	})

	it("passes and opts in for the node-invoker if the NPM app BP applied", func() {
		f := test.NewEnvironmentFactory(t)
		defer f.Restore()

		f.Console.In(t, fmt.Sprintf("[%s]", npmdetect.NPMDependency))

		if err := libjavabuildpack.WriteToFile(strings.NewReader(`artifact = "my.js"`), filepath.Join(f.Application, "riff.toml"), 0644); err != nil {
			t.Fatal(err)
		}

		main()

		if *f.ExitStatus != 0 {
			t.Errorf("os.Exit = %d, expected 0", *f.ExitStatus)
		}

		if bp, err := libbuildpack.NewBuildPlan(strings.NewReader(f.Console.Out(t)), libbuildpack.NewLogger(nil, nil)) ; err != nil {
			t.Fatal(err)
		} else {
			if j, ok := bp[node.RiffNodeInvokerDependency] ; !ok {
				t.Error("expected the node invoker to be added: ", bp)
			} else {
				if j.Metadata[node.FunctionArtifact] != "my.js" {
					t.Error("fn key not set: ", bp)
				}
			}
		}
	})

	it("passes and opts in for the node-invoker if the NPM app BP did not apply, but artifact is .js", func() {
		f := test.NewEnvironmentFactory(t)
		defer f.Restore()

		f.Console.In(t, "") // empty BP

		if err := libjavabuildpack.WriteToFile(strings.NewReader(`module.exports = x => x**2`), filepath.Join(f.Application, "my.js"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := libjavabuildpack.WriteToFile(strings.NewReader(`artifact = "my.js"`), filepath.Join(f.Application, "riff.toml"), 0644); err != nil {
			t.Fatal(err)
		}

		main()

		if *f.ExitStatus != 0 {
			t.Errorf("os.Exit = %d, expected 0", *f.ExitStatus)
		}

		if bp, err := libbuildpack.NewBuildPlan(strings.NewReader(f.Console.Out(t)), libbuildpack.NewLogger(nil, nil)) ; err != nil {
			t.Fatal(err)
		} else {
			if j, ok := bp[node.RiffNodeInvokerDependency] ; !ok {
				t.Error("expected the node invoker to be added: ", bp)
			} else {
				if j.Metadata[node.FunctionArtifact] != "my.js" {
					t.Error("fn key not set: ", bp)
				}
			}
		}
	})

	it("passes and opts in for the command-invoker if the artifact is executable", func() {
		f := test.NewEnvironmentFactory(t)
		defer f.Restore()

		f.Console.In(t, "") // empty BP

		if err := libjavabuildpack.WriteToFile(strings.NewReader(`some bash`), filepath.Join(f.Application, "fn.sh"), 0744/*<-executable*/); err != nil {
			t.Fatal(err)
		}
		if err := libjavabuildpack.WriteToFile(strings.NewReader(`artifact = "fn.sh"`), filepath.Join(f.Application, "riff.toml"), 0644); err != nil {
			t.Fatal(err)
		}

		main()

		if *f.ExitStatus != 0 {
			t.Errorf("os.Exit = %d, expected 0", *f.ExitStatus)
		}

		if bp, err := libbuildpack.NewBuildPlan(strings.NewReader(f.Console.Out(t)), libbuildpack.NewLogger(nil, nil)) ; err != nil {
			t.Fatal(err)
		} else {
			if j, ok := bp[command.RiffCommandInvokerDependency] ; !ok {
				t.Error("expected the command invoker to be added: ", bp)
			} else {
				if j.Metadata[command.Command] != "fn.sh" {
					t.Error("command key not set: ", bp)
				}
			}
		}
	})

	it("fails if ambiguity", func() {
		f := test.NewEnvironmentFactory(t)
		defer f.Restore()

		f.Console.In(t, fmt.Sprintf("[%s]\n[%s]", jvm_application_buildpack.JVMApplication, npmdetect.NPMDependency))

		if err := libjavabuildpack.WriteToFile(strings.NewReader(`some bash`), filepath.Join(f.Application, "fn.sh"), 0744/*<-executable*/); err != nil {
			t.Fatal(err)
		}
		if err := libjavabuildpack.WriteToFile(strings.NewReader(`artifact = "fn.sh"`), filepath.Join(f.Application, "riff.toml"), 0644); err != nil {
			t.Fatal(err)
		}

		main()

		if *f.ExitStatus != Error_DetectAmbiguity {
			t.Errorf("os.Exit = %d, expected %d", *f.ExitStatus, Error_DetectAmbiguity)
		}

	})

	it("override resolves ambiguity", func() {
		f := test.NewEnvironmentFactory(t)
		defer f.Restore()

		f.Console.In(t, fmt.Sprintf("[%s]\n[%s]", jvm_application_buildpack.JVMApplication, npmdetect.NPMDependency))

		if err := libjavabuildpack.WriteToFile(strings.NewReader(`some bash`), filepath.Join(f.Application, "fn.sh"), 0744/*<-executable*/); err != nil {
			t.Fatal(err)
		}
		if err := libjavabuildpack.WriteToFile(strings.NewReader("artifact = \"fn.sh\"\noverride = \"java\""), filepath.Join(f.Application, "riff.toml"), 0644); err != nil {
			t.Fatal(err)
		}

		main()

		if *f.ExitStatus != 0 {
			t.Errorf("os.Exit = %d, expected 0", *f.ExitStatus)
		}

		if bp, err := libbuildpack.NewBuildPlan(strings.NewReader(f.Console.Out(t)), libbuildpack.NewLogger(nil, nil)) ; err != nil {
			t.Fatal(err)
		} else {
			if _, ok := bp[java.RiffInvokerDependency] ; !ok {
				t.Error("expected the java invoker to be added: ", bp)
			}
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
