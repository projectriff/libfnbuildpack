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
	"github.com/cloudfoundry/jvm-application-buildpack"
	"github.com/cloudfoundry/libjavabuildpack"
	npmdetect "github.com/cloudfoundry/npm-cnb/detect"
	"github.com/projectriff/riff-buildpack"
	"github.com/projectriff/riff-buildpack/command"
	"github.com/projectriff/riff-buildpack/java"
	"github.com/projectriff/riff-buildpack/node"
	"os"
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
	detect, err := libjavabuildpack.DefaultDetect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize Detect: %s\n", err.Error())
		os.Exit(Error_Initialize)
	}

	metadata, ok, err := riff_buildpack.NewMetadata(detect.Application, detect.Logger)
	if err != nil {
		detect.Logger.Info("Unable to read riff metadata: %s", err.Error())
		detect.Error(Error_ReadMetadata)
		return
	}

	if !ok {
		detect.Fail()
		return
	}

	detected := []string{}

	if metadata.Override == "" {
		// Try java
		if _, ok := detect.BuildPlan[jvm_application_buildpack.JVMApplication]; ok {
			detected = append(detected, "java")
		}

		// Try npm
		if _, ok := detect.BuildPlan[npmdetect.NPMDependency]; ok {
			detected = append(detected, "node")
		} else {
			// Try node
			if ok, err := node.DetectNode(detect, metadata); err != nil {
				detect.Logger.Info("Error trying to use node invoker: %s", err.Error())
				detect.Error(Error_DetectInternalError)
				return
			} else if ok {
				detected = append(detected, "node")
			}
		}

		// Try command invoker as last resort
		if ok, err := command.DetectCommand(detect, metadata); err != nil {
			detect.Logger.Info("Error trying to use command invoker: %s", err.Error())
			detect.Error(Error_DetectInternalError)
			return
		} else if ok {
			detected = append(detected, "command")
		}

		if len(detected) == 0 {
			detect.Logger.Info("Detected riff function but unable to determine function type.")
			detect.Error(Error_DetectedNone)
			return
		} else if len(detected) > 1 {
			detect.Logger.Info("Detected riff function but ambiguous language detected: %v.", detected)
			detect.Error(Error_DetectAmbiguity)
			return
		}

		detect.Logger.Debug("Detected language: %q.", detected[0])

	} else {
		detected = []string{metadata.Override}
	}

	switch detected[0] {
	case "java":
		detect.Pass(java.BuildPlanContribution(metadata))
		return
	case "node":
		detect.Pass(node.BuildPlanContribution(metadata))
		return
	case "command":
		detect.Pass(command.BuildPlanContribution(metadata))
		return
	default:
		detect.Logger.Info("Unsupported language: %v.", detected[0])
		detect.Error(Error_UnsupportedLanguage)
		return
	}
}
