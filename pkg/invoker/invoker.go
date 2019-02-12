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
 */

package invoker

import (
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/projectriff/riff-buildpack/pkg/metadata"
)

const (
	Error_Initialize          = 101
	Error_ReadMetadata        = 102
	Error_DetectedNone        = 103
	Error_DetectAmbiguity     = 104
	Error_UnsupportedLanguage = 105
	Error_DetectInternalError = 106
)

type RiffBuildpackContribution struct {
	Name      string
	Detect    func(detect detect.Detect, metadata metadata.Metadata) (bool, int, error)
	Invoker   func(build build.Build) (RiffInvoker, bool, error)
	BuildPlan func(detect detect.Detect, metadata metadata.Metadata) buildplan.BuildPlan
}

type RiffInvoker interface {
	Contribute() error
}
