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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContainer(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()

	t.Run("not-supported-container", func(t *testing.T) {
		cnt, err := NewContainer(&ctx, "test_drv", "")
		assert.Equal(nil, cnt)
		assert.Equal(ErrContainerDrvNotSupported{Driver: "test_drv"}, err)
	})

	t.Run("empty-container", func(t *testing.T) {
		cnt, err := NewContainer(&ctx, "", "")
		assert.Equal(nil, cnt)
		assert.Equal(ErrNoContainerDriver{}, err)
	})
}
