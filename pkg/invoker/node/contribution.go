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

package node

import (
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/npm-cnb/modules"
	"github.com/projectriff/riff-buildpack/pkg/invoker"
	"github.com/projectriff/riff-buildpack/pkg/metadata"
)

var RiffBuildpackContribution = invoker.RiffBuildpackContribution{
	Name: "node",
	Detect: func(detect detect.Detect, metadata metadata.Metadata) (bool, int, error) {
		// Try npm
		if _, ok := detect.BuildPlan[modules.Dependency]; ok {
			return ok, 0, nil
		} else {
			// Try node
			if ok, err := DetectNode(detect, metadata); err != nil {
				detect.Logger.Info("Error trying to use node invoker: %s", err.Error())
				return false, detect.Error(invoker.Error_DetectInternalError), nil
			} else {
				return ok, 0, nil
			}
		}
	},
	Invoker:   NewNodeInvoker,
	BuildPlan: BuildPlanContribution,
}
