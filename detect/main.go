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

	"github.com/cloudfoundry/jvm-application-buildpack/jvmapplication"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/npm-cnb/modules"
	"github.com/projectriff/riff-buildpack/command"
	"github.com/projectriff/riff-buildpack/java"
	"github.com/projectriff/riff-buildpack/metadata"
	"github.com/projectriff/riff-buildpack/node"
)

const (
	Error_Initialize          = 101
	Error_ReadMetadata        = 102
	Error_DetectedNone        = 103
	Error_DetectAmbiguity     = 104
	Error_UnsupportedLanguage = 105
	Error_DetectInternalError = 106
)

func main() {
	detect, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Detect: %s\n", err)
		os.Exit(Error_Initialize)
	}

	if err := detect.BuildPlan.Init(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Build Plan: %s\n", err)
		os.Exit(Error_Initialize)
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
		return detect.Error(Error_ReadMetadata), fmt.Errorf("unable to read riff metadata: %s", err.Error())
	}

	if !ok {
		return detect.Fail(), nil
	}

	detected := []string{}

	if metadata.Override == "" {
		// Try java
		if _, ok := detect.BuildPlan[jvmapplication.Dependency]; ok {
			detected = append(detected, "java")
		}

		// Try npm
		if _, ok := detect.BuildPlan[modules.Dependency]; ok {
			detected = append(detected, "node")
		} else {
			// Try node
			if ok, err := node.DetectNode(detect, metadata); err != nil {
				detect.Logger.Info("Error trying to use node invoker: %s", err.Error())
				return detect.Error(Error_DetectInternalError), nil
			} else if ok {
				detected = append(detected, "node")
			}
		}

		// Try command invoker as last resort
		if ok, err := command.DetectCommand(detect, metadata); err != nil {
			return detect.Error(Error_DetectInternalError), fmt.Errorf("error trying to use command invoker: %s", err.Error())
		} else if ok {
			detected = append(detected, "command")
		}

		if len(detected) == 0 {
			return detect.Error(Error_DetectedNone), fmt.Errorf("detected riff function but unable to determine function type")
		} else if len(detected) > 1 {
			return detect.Error(Error_DetectAmbiguity), fmt.Errorf("detected riff function but ambiguous language detected: %v", detected)
		}

		detect.Logger.Debug("Detected language: %q.", detected[0])
	} else {
		detected = []string{metadata.Override}
	}

	switch detected[0] {
	case "java":
		return detect.Pass(java.BuildPlanContribution(detect, metadata))
	case "node":
		return detect.Pass(node.BuildPlanContribution(detect, metadata))
	case "command":
		return detect.Pass(command.BuildPlanContribution(detect, metadata))
	default:
		return detect.Error(Error_UnsupportedLanguage), fmt.Errorf("unsupported language: %v", detected[0])
	}
}
