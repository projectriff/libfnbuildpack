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

	"github.com/buildpack/libbuildpack"
	"github.com/projectriff/riff-buildpack"
	"github.com/projectriff/riff-buildpack/java"
)

func main() {
	detect, err := libbuildpack.DefaultDetect()
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

	// TODO use constants for jvm-application
	if _, ok := detect.BuildPlan["jvm-application"]; ok {
		detect.Logger.Debug("Riff Java application")
		detect.Pass(java.RiffInvoker{}.BuildPlanContribution(metadata))
		return
	}

	detect.Logger.Info("Detected Riff application but unable to determine application type.")
	detect.Error(103)
	return
}
