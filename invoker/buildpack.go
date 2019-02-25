/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package invoker

import (
	"fmt"
	"os"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/projectriff/riff-buildpack/metadata"
)

const (
	Error_Initialize          = 101
	Error_ReadMetadata        = 102
	Error_DetectedNone        = 103
	Error_DetectAmbiguity     = 104
	Error_UnsupportedLanguage = 105
	Error_DetectInternalError = 106
)

type Buildpack interface {
	Name() string
	Detect(detect detect.Detect, metadata metadata.Metadata) (bool, error)
	BuildPlan(detect detect.Detect, metadata metadata.Metadata) buildplan.BuildPlan
	Invoker(build build.Build) (Invoker, bool, error)
}

type Invoker interface {
	Contribute() error
}

func Detect(bp Buildpack) {
	detect, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Detect: %s\n", err)
		os.Exit(Error_Initialize)
	}

	if err := detect.BuildPlan.Init(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Build Plan: %s\n", err)
		os.Exit(Error_Initialize)
	}

	if code, err := doDetect(bp, detect); err != nil {
		detect.Logger.Info(err.Error())
		os.Exit(code)
	} else {
		os.Exit(code)
	}
}

func doDetect(bp Buildpack, detect detect.Detect) (int, error) {
	metadata, ok, err := metadata.NewMetadata(detect.Application, detect.Logger)
	if err != nil {
		return detect.Error(Error_ReadMetadata), fmt.Errorf("unable to read riff metadata: %s", err.Error())
	}

	if !ok {
		return detect.Fail(), nil
	}

	detected := false

	if metadata.Override != "" {
		if metadata.Override == bp.Name() {
			detected = true
			detect.Logger.Debug("Override language: %q.", bp.Name())
		}
	} else {
		if detected, err = bp.Detect(detect, metadata); err != nil {
			detect.Logger.Info("Error trying to use %s invoker: %s", bp.Name(), err.Error())
			return detect.Error(Error_DetectInternalError), nil
		}

		if detected {
			detect.Logger.Debug("Detected language: %q.", bp.Name())
		}
	}

	if detected {
		return detect.Pass(bp.BuildPlan(detect, metadata))
	}

	return detect.Fail(), nil
}

func Build(bp Buildpack) {
	build, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Build: %s\n", err)
		os.Exit(101)
	}

	if code, err := doBuild(bp, build); err != nil {
		build.Logger.Info(err.Error())
		os.Exit(code)
	} else {
		os.Exit(code)
	}
}

func doBuild(bp Buildpack, build build.Build) (int, error) {
	build.Logger.FirstLine(build.Logger.PrettyIdentity(build.Buildpack))

	if invoker, ok, err := bp.Invoker(build); err != nil {
		return build.Failure(105), err
	} else if ok {
		if err = invoker.Contribute(); err != nil {
			return build.Failure(106), err
		}
		return build.Success(buildplan.BuildPlan{})
	}

	build.Logger.Info("Buildpack passed detection but did not know how to actually build. Should never happen.")
	return build.Failure(104), nil
}
