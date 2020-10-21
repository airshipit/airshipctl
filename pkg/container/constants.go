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

const (
	// CreatedContainerStatus indicates container has been created
	CreatedContainerStatus = "created"
	// RunningContainerStatus indicates container is in running state
	RunningContainerStatus = "running"
	// PausedContainerStatus indicates container is in paused state
	PausedContainerStatus = "paused"
	// RestartedContainerStatus indicates container has re-started
	RestartedContainerStatus = "restarted"
	// RemovingContainerStatus indicates container is being removed
	RemovingContainerStatus = "removing"
	// ExitedContainerStatus indicates container has exited
	ExitedContainerStatus = "exited"
	// DeadContainerStatus indicates container is dead
	DeadContainerStatus = "dead"
)
