package implementations

import (
	"fmt"
)

// ErrVersionNotDefined is returned when requested version is not present in repository
type ErrVersionNotDefined struct {
	Version string
}

func (e ErrVersionNotDefined) Error() string {
	return fmt.Sprintf(`version %s is not defined in the repository`, e.Version)
}

// ErrNoVersionsAvailable is returned when version map is empty or not defined
type ErrNoVersionsAvailable struct {
	Versions map[string]string
}

func (e ErrNoVersionsAvailable) Error() string {
	return fmt.Sprintf(`version map is empty or not defined, %v`, e.Versions)
}
