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

package command

import (
	"fmt"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/projectriff/riff-buildpack/pkg/invoker"
	"github.com/projectriff/riff-buildpack/pkg/metadata"
)

var RiffBuildpackContribution = invoker.RiffBuildpackContribution{
	Name: "command",
	Detect: func(detect detect.Detect, metadata metadata.Metadata) (bool, int, error) {
		if ok, err := DetectCommand(detect, metadata); err != nil {
			return false, detect.Error(invoker.Error_DetectInternalError), fmt.Errorf("error trying to use command invoker: %s", err.Error())
		} else {
			return ok, 0, nil
		}
	},
	Invoker:   NewCommandInvoker,
	BuildPlan: BuildPlanContribution,
}
