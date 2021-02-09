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

package container

import (
	"io"

	"opendev.org/airship/airshipctl/pkg/container"
)

// MockContainer implements Container for unit test purposes
type MockContainer struct {
	MockImagePull         func() error
	MockRunCommand        func() error
	MockGetContainerLogs  func() (io.ReadCloser, error)
	MockRmContainer       func() error
	MockGetID             func() string
	MockWaitUntilFinished func() error
	MockInspectContainer  func() (container.State, error)
}

// ImagePull Container interface implementation for unit test purposes
func (mc *MockContainer) ImagePull() error {
	return mc.MockImagePull()
}

// RunCommand Container interface implementation for unit test purposes
func (mc *MockContainer) RunCommand(container.RunCommandOptions) error {
	return mc.MockRunCommand()
}

// GetContainerLogs Container interface implementation for unit test purposes
func (mc *MockContainer) GetContainerLogs(container.GetLogOptions) (io.ReadCloser, error) {
	return mc.MockGetContainerLogs()
}

// RmContainer Container interface implementation for unit test purposes
func (mc *MockContainer) RmContainer() error {
	return mc.MockRmContainer()
}

// GetID Container interface implementation for unit test purposes
func (mc *MockContainer) GetID() string {
	return mc.MockGetID()
}

// WaitUntilFinished Container interface implementation for unit test purposes
func (mc *MockContainer) WaitUntilFinished() error {
	return mc.MockWaitUntilFinished()
}

// InspectContainer Container interface implementation for unit test purposes
func (mc *MockContainer) InspectContainer() (container.State, error) {
	return mc.MockInspectContainer()
}
