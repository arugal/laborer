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

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/antihax/optional"
	"github.com/scultura-org/harborapi"
)

type NotFoundRepoError struct {
	message string
}

func (e NotFoundRepoError) Error() string {
	return e.message
}

// RepositoryService 镜像服务
type RepositoryService interface {
	// 获取镜像最新的 tag
	LatestTag(projectName, repoName string) (tag string, err error)
}

// RepositoryServiceOption 设置 repository service
type RepositoryServiceOption func(service *harborRepositoryService)

// WithBasePath
func WithBasePath(basePath string) RepositoryServiceOption {
	return func(service *harborRepositoryService) {
		service.basePath = basePath
	}
}

// WithInsecureSkipVerify
func WithInsecureSkipVerify(insecureSkipVerify bool) RepositoryServiceOption {
	return func(service *harborRepositoryService) {
		service.insecureSkipVerify = insecureSkipVerify
	}
}

// NewRepositoryService
func NewRepositoryService(opts ...RepositoryServiceOption) (RepositoryService, error) {
	service := &harborRepositoryService{}

	// with options
	for _, opt := range opts {
		opt(service)
	}

	if service.basePath == "" {
		return nil, errors.New("baseURL must be set.")
	}

	cfg := harborapi.NewConfiguration()
	cfg.BasePath = service.basePath

	// insecureSkipVerify
	if service.insecureSkipVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		cfg.HTTPClient = &http.Client{
			Transport: tr,
		}
	}

	service.apiClient = harborapi.NewAPIClient(cfg)

	return service, nil
}

type harborRepositoryService struct {
	basePath           string
	insecureSkipVerify bool

	apiClient *harborapi.APIClient
}

func (h *harborRepositoryService) LatestTag(projectName, repoName string) (tag string, err error) {
	artifacts, _, err := h.apiClient.ArtifactApi.ListArtifacts(context.Background(), projectName, repoName, &harborapi.ArtifactApiListArtifactsOpts{
		PageSize: optional.NewInt64(100),
	})
	if err != nil {
		return
	}
	if len(artifacts) == 0 {
		return tag, &NotFoundRepoError{message: fmt.Sprintf("repo %s/%s not found.", projectName, repoName)}
	}

	// 根据 PushTime 对 artifact 和 tag 排序, 找出最新 push 的 image
	latestArtifact := ArtifactSlice(artifacts).Sort().Latest()
	latestTag := TagSlice(latestArtifact.Tags).Sort().Latest()

	return latestTag.Name, nil
}

// 根据 PushTime 对 Artifact 排序
type ArtifactSlice []harborapi.Artifact

func (t ArtifactSlice) Len() int           { return len(t) }
func (t ArtifactSlice) Less(i, j int) bool { return t[i].PushTime.Before(t[j].PushTime) }
func (t ArtifactSlice) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

func (t ArtifactSlice) Sort() ArtifactSlice {
	sort.Sort(t)
	return t
}

func (t ArtifactSlice) Latest() harborapi.Artifact {
	return t[len(t)-1]
}

// 根据 PushTime 对 Tag 排序
type TagSlice []harborapi.Tag

func (t TagSlice) Len() int           { return len(t) }
func (t TagSlice) Less(i, j int) bool { return t[i].PushTime.Before(t[j].PushTime) }
func (t TagSlice) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

func (t TagSlice) Sort() TagSlice {
	sort.Sort(t)
	return t
}

func (t TagSlice) Latest() harborapi.Tag {
	return t[len(t)-1]
}
