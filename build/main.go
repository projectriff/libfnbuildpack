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

	"github.com/cloudfoundry/libjavabuildpack"
	"github.com/projectriff/riff-buildpack/java"
)

func main() {
	build, err := libjavabuildpack.DefaultBuild()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize Build: %s\n", err.Error())
		os.Exit(101)
	}

	build.Logger.FirstLine(build.Logger.PrettyVersion(build.Buildpack))

	if invoker, ok, err := java.NewRiffInvoker(build); err != nil {
		build.Logger.Info(err.Error())
		build.Failure(102)
		return
	} else if ok {
		if err = invoker.Contribute(); err != nil {
			build.Logger.Info(err.Error())
			build.Failure(103)
			return
		}
	}

	build.Success()
}
