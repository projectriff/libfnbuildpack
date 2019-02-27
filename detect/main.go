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
	"os"
	"strings"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/projectriff/riff-buildpack/function"
)

const (
	Dependency    = "riff-buildpack"
	invokerPrefix = "riff-invoker-"
)

func main() {
	detect, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Detect: %s\n", err)
		os.Exit(function.Error_Initialize)
	}

	if err := detect.BuildPlan.Init(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Build Plan: %s\n", err)
		os.Exit(function.Error_Initialize)
	}

	if code, err := d(detect); err != nil {
		detect.Logger.Info(err.Error())
		os.Exit(code)
	} else {
		os.Exit(code)
	}
}

func d(detect detect.Detect) (int, error) {
	_, ok, err := function.NewMetadata(detect.Application, detect.Logger)
	if err != nil {
		return detect.Error(function.Error_DetectReadMetadata), fmt.Errorf("unable to read riff metadata: %s", err.Error())
	}

	if !ok {
		return detect.Fail(), nil
	}

	detected := []string{}

	for name := range detect.BuildPlan {
		if strings.HasPrefix(name, invokerPrefix) {
			detected = append(detected, name[len(invokerPrefix):])
		}
	}

	if len(detected) == 0 {
		return detect.Error(function.Error_DetectedNone), fmt.Errorf("detected riff function but unable to determine function type")
	} else if len(detected) > 1 {
		return detect.Error(function.Error_DetectAmbiguity), fmt.Errorf("detected riff function but ambiguous language detected: %v", detected)
	}

	return detect.Pass(buildplan.BuildPlan{})
}
