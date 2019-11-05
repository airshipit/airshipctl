package container

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContainer(t *testing.T) {
	ctx := context.Background()
	_, actualErr := NewContainer(&ctx, "test_drv", "")
	expectedErr := ErrContainerDrvNotSupported{Driver: "test_drv"}
	errS := fmt.Sprintf(
		"Call NewContainer should have returned error %s, got %s",
		expectedErr,
		actualErr,
	)
	assert.Equal(t, actualErr, expectedErr, errS)
}
