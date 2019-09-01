package cloudinit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/testutil"
)

func TestGetCloudData(t *testing.T) {

	fSys := testutil.SetupTestFs(t, "testdata")
	bundle, err := document.NewBundle(fSys, "/", "/")
	require.NoError(t, err, "Building Bundle Failed")

	tests := []struct {
		ann              string
		expectedUserData []byte
		expectedNetData  []byte
		expectedErr      error
	}{
		{
			ann:              "test=test",
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: document.ErrDocNotFound{
				Annotation: "test=test",
				Kind:       "Secret",
			},
		},
		{
			ann:              "airshipit.org/clustertype=nodata",
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: ErrDataNotSupplied{
				DocName: "node1-bmc-secret1",
				Key:     "netconfig",
			},
		},
		{
			ann:              "test=nodataforcfg",
			expectedUserData: nil,
			expectedNetData:  nil,
			expectedErr: ErrDataNotSupplied{
				DocName: "node1-bmc-secret2",
				Key:     "netconfig",
			},
		},
		{
			ann:              "airshipit.org/clustertype=ephemeral",
			expectedUserData: []byte("cloud-init"),
			expectedNetData:  []byte("netconfig\n"),
			expectedErr:      nil,
		},
	}

	for _, tt := range tests {
		actualUserData, actualNetData, actualErr := GetCloudData(bundle, tt.ann)

		assert.Equal(t, tt.expectedUserData, actualUserData)
		assert.Equal(t, tt.expectedNetData, actualNetData)
		assert.Equal(t, tt.expectedErr, actualErr)
	}
}
