/*
 Copyright 2020 arugal.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package repository

import "fmt"

type mockRepositoryService struct {
	tags map[string]string
}

func (m mockRepositoryService) LatestTag(host, projectName, repoName string) (tag string, err error) {
	var image string
	if host != "" && projectName != "" {
		image = fmt.Sprintf("%s/%s/%s", host, projectName, repoName)
	} else if projectName != "" {
		image = fmt.Sprintf("%s/%s", projectName, repoName)
	} else {
		image = repoName
	}

	if tag, ok := m.tags[image]; ok {
		return tag, err
	}
	return tag, &NotFoundRepoError{message: fmt.Sprintf("repo %s/%s not found.", projectName, repoName)}
}
