/*
 * Copyright 2019 the original author or authors.
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

package java

import (
	"github.com/cloudfoundry/jvm-application-buildpack/jvmapplication"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/projectriff/riff-buildpack/pkg/invoker"
	"github.com/projectriff/riff-buildpack/pkg/metadata"
)

var RiffBuildpackContribution = invoker.RiffBuildpackContribution{
	Name: "java",
	Detect: func(detect detect.Detect, metadata metadata.Metadata) (bool, int, error) {
		_, ok := detect.BuildPlan[jvmapplication.Dependency]
		return ok, 0, nil
	},
	Invoker:   NewJavaInvoker,
	BuildPlan: BuildPlanContribution,
}
