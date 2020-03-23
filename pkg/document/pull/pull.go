package pull

import (
	"opendev.org/airship/airshipctl/pkg/document/repo"
	"opendev.org/airship/airshipctl/pkg/environment"
)

type Settings struct {
	*environment.AirshipCTLSettings
}

func (s *Settings) Pull() error {
	err := s.cloneRepositories()
	if err != nil {
		return err
	}

	return nil
}

func (s *Settings) cloneRepositories() error {
	// Clone main repository
	currentManifest, err := s.Config.CurrentContextManifest()
	if err != nil {
		return err
	}

	// Clone repositories
	for _, extraRepoConfig := range currentManifest.Repositories {
		repository, err := repo.NewRepository(currentManifest.TargetPath, extraRepoConfig)
		if err != nil {
			return err
		}
		err = repository.Download(extraRepoConfig.ToCheckoutOptions(true).Force)
		if err != nil {
			return err
		}
		repository.Driver.Close()
	}

	return nil
}
