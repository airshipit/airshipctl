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

// ErrDuplicateNetworkDataDocuments error returned if multiple network documents
// were found with the same name in the same namespace
type ErrDuplicateNetworkDataDocuments struct {
	DocName   string
	Namespace string
}

func (e ErrDataNotSupplied) Error() string {
	return fmt.Sprintf("Document %s has no key %s", e.DocName, e.Key)
}

func (e ErrDuplicateNetworkDataDocuments) Error() string {
	return fmt.Sprintf("Found more than one document with the name %s in namespace %s", e.DocName, e.Namespace)
}
