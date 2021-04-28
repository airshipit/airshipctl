/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package util

import (
	"os"

	"opendev.org/airship/airshipctl/pkg/log"
)

// EnvVar is an object to store key:value pairs to set/unset environment variables
type EnvVar struct {
	Key   string
	Value string
}

// Setenv is used to set variable number of env variables, if an error occurs it will be printed to logger
func Setenv(envVars ...EnvVar) {
	for _, envVar := range envVars {
		if err := os.Setenv(envVar.Key, envVar.Value); err != nil {
			log.Printf("unable to set '%s' env variable, reason '%s'", envVar.Key, err.Error())
		}
	}
}

// Unsetenv is used to unset variable number of env variables, if an error occurs it will be printed to logger
func Unsetenv(envVars ...EnvVar) {
	for _, envVar := range envVars {
		if err := os.Unsetenv(envVar.Key); err != nil {
			// Unsetenv on Unix never returns an error [1], so we can't emulate it in unit tests and
			// this error message will never be printed
			// [1] https://golang.org/src/syscall/env_unix.go
			log.Printf("unable to unset '%s' env variable, reason '%s'", envVar.Key, err.Error())
		}
	}
}
