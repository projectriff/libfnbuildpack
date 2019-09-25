/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package function

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/detect"
)

const (
	Error_Initialization          = 101
	Error_ComponentInitialization = 102
	Error_ComponentContribution   = 103
	Error_ComponentInternal       = 104
	Error_ReadMetadata            = 105
)

type Buildpack struct {
	BuildpackImplementation BuildpackImplementation
}

func (b Buildpack) Build() {
	build, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Build: %s\n", err)
		os.Exit(Error_Initialization)
	}

	if code, err := b.doBuild(build); err != nil {
		build.Logger.Info(err.Error())
		os.Exit(code)
	} else {
		os.Exit(code)
	}
}

func (b Buildpack) doBuild(build build.Build) (int, error) {
	build.Logger.Title(build.Buildpack)

	if code, err := b.BuildpackImplementation.Build(build); err != nil {
		return build.Failure(code), fmt.Errorf("unable to build invoker %q: %s", b.BuildpackImplementation.Id(), err)
	}
	return build.Success()
}

func (b Buildpack) Detect() {
	detect, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Detect: %s\n", err)
		os.Exit(Error_Initialization)
	}

	if code, err := b.doDetect(detect); err != nil {
		detect.Logger.Info(err.Error())
		os.Exit(code)
	} else {
		os.Exit(code)
	}
}

func (b Buildpack) doDetect(d detect.Detect) (int, error) {
	m, ok, err := NewMetadata(d.Application, d.Logger)
	if err != nil {
		return d.Error(Error_ReadMetadata), fmt.Errorf("unable to read riff metadata: %s", err.Error())
	}

	if !ok {
		return d.Fail(), nil
	}

	if m.Override != "" && m.Override != b.BuildpackImplementation.Id() {
		return d.Fail(), nil
	}

	code, err := b.BuildpackImplementation.Detect(d, m)
	if err != nil {
		d.Logger.Info("Error trying to use %s invoker: %s", b.BuildpackImplementation.Id(), err.Error())
		return d.Error(code), err
	}

	if code == detect.PassStatusCode {
		d.Logger.Debug("Detected language: %q.", b.BuildpackImplementation.Id())
	}

	return code, err
}
