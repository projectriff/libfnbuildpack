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

	"github.com/cloudfoundry/libcfbuildpack/detect"
	buildpack "github.com/projectriff/riff-buildpack/pkg"
	"github.com/projectriff/riff-buildpack/pkg/invoker"
	"github.com/projectriff/riff-buildpack/pkg/metadata"
)

func main() {
	detect, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Detect: %s\n", err)
		os.Exit(invoker.Error_Initialize)
	}

	if err := detect.BuildPlan.Init(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Build Plan: %s\n", err)
		os.Exit(invoker.Error_Initialize)
	}

	if code, err := d(detect); err != nil {
		detect.Logger.Info(err.Error())
		os.Exit(code)
	} else {
		os.Exit(code)
	}
}

func d(detect detect.Detect) (int, error) {
	metadata, ok, err := metadata.NewMetadata(detect.Application, detect.Logger)
	if err != nil {
		return detect.Error(invoker.Error_ReadMetadata), fmt.Errorf("unable to read riff metadata: %s", err.Error())
	}

	if !ok {
		return detect.Fail(), nil
	}

	detected := []string{}

	if metadata.Override == "" {
		for _, contribution := range buildpack.RiffBuildpackContributions {
			ok, code, err := contribution.Detect(detect, metadata)
			if code != 0 {
				return code, err
			}
			if ok {
				detected = append(detected, contribution.Name)
			}
		}

		if len(detected) == 0 {
			return detect.Error(invoker.Error_DetectedNone), fmt.Errorf("detected riff function but unable to determine function type")
		} else if len(detected) > 1 {
			return detect.Error(invoker.Error_DetectAmbiguity), fmt.Errorf("detected riff function but ambiguous language detected: %v", detected)
		}

		detect.Logger.Debug("Detected language: %q.", detected[0])
	} else {
		detected = []string{metadata.Override}
	}

	for _, contribution := range buildpack.RiffBuildpackContributions {
		if detected[0] == contribution.Name {
			return detect.Pass(contribution.BuildPlan(detect, metadata))
		}
	}
	return detect.Error(invoker.Error_UnsupportedLanguage), fmt.Errorf("unsupported language: %v", detected[0])
}
