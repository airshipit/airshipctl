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

package container_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/pkg/container"
)

func TestNewContainer(t *testing.T) {
	a := assert.New(t)

	ctx := context.Background()

	t.Run("not-supported-container", func(t *testing.T) {
		cnt, err := container.NewContainer(ctx, "test_drv", "")
		a.Equal(nil, cnt)
		a.Equal(container.ErrContainerDrvNotSupported{Driver: "test_drv"}, err)
	})

	t.Run("empty-container", func(t *testing.T) {
		cnt, err := container.NewContainer(ctx, "", "")
		a.Equal(nil, cnt)
		a.Equal(container.ErrNoContainerDriver{}, err)
	})
}
