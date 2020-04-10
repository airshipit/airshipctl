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
	currentManifest, err := s.Config().CurrentContextManifest()
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
