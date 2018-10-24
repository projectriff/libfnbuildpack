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
	"github.com/projectriff/riff-buildpack"
	"github.com/projectriff/riff-buildpack/command"
	"github.com/projectriff/riff-buildpack/java"
	"os"
)

func main() {
	detect, err := libjavabuildpack.DefaultDetect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize Detect: %s\n", err.Error())
		os.Exit(101)
	}

	metadata, ok, err := riff_buildpack.NewMetadata(detect.Application, detect.Logger)
	if err != nil {
		detect.Logger.Info("Unable to read riff metadata: %s", err.Error())
		detect.Error(102)
		return
	}

	if !ok {
		detect.Fail()
		return
	}

	if _, ok := detect.BuildPlan[jvm_application_buildpack.JVMApplication]; ok {
		detect.Logger.Debug("riff Java application")
		detect.Pass(java.BuildPlanContribution(metadata))
		return
	} else {
		// Try command invoker as last resort
		if ok, err := command.DetectCommand(detect, metadata) ; err != nil {
			detect.Logger.Info("Error trying to use command invoker: %s", err.Error())
			detect.Error(104)
			return
		} else if ok {
			detect.Logger.Debug("riff Command application")
			detect.Pass(command.BuildPlanContribution(metadata))
			return
		}
	}

	detect.Logger.Info("Detected riff application but unable to determine application type.")
	detect.Error(103)
	return
}
