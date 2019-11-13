package cloudinit

import (
	"fmt"
)

// ErrDataNotSupplied error returned of no user-data or network configuration
// in the Secret
type ErrDataNotSupplied struct {
	DocName string
	Key     string
}

func (e ErrDataNotSupplied) Error() string {
	return fmt.Sprintf("Document %s has no key %s", e.DocName, e.Key)
}
