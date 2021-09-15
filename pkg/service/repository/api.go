/*
 Copyright 2021 zhangwei24@apache.org

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
	"fmt"
	ht "net/http"

	"github.com/antihax/optional"
	"github.com/scultura-org/harborapi"
	"k8s.io/klog"
)

// NotFoundRepoError 未找到对应的 repo
type NotFoundRepoError struct {
	message string
}

func (e NotFoundRepoError) Error() string {
	return e.message
}

// NotSupportRegisterError 不支持的注册中心地址
type NotSupportRegisterError struct {
	host string
}

func (e NotSupportRegisterError) Error() string {
	return fmt.Sprintf("only support %s", e.host)
}

// RepositoryService 镜像服务
type RepositoryService interface {
	// 获取镜像最新的 tag
	LatestTag(host, projectName, repoName string) (tag string, err error)
}

// TODO support more repository
// NewRepositoryService
func NewRepositoryService(options *RepositoryServiceOptions) (RepositoryService, error) {
	if options.Mock {
		// mock service
		return &mockRepositoryService{
			tags: options.MockTags,
		}, nil
	}
	if options.Host == "" {
		return &ignoreRepositoryService{}, nil
	}

	if options.ApiPathPrefix == "" {
		klog.Warning("repository service api path prefix is empty")
	}

	cfg := harborapi.NewConfiguration()
	cfg.BasePath = fmt.Sprintf("%s://%s%s", options.Protocol, options.Host, options.ApiPathPrefix)

	// insecureSkipVerify
	if options.InsecureSkipVerify && options.Protocol == "https" {
		tr := &ht.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		cfg.HTTPClient = &ht.Client{
			Transport: tr,
		}
	}

	service := &harborRepositoryService{
		host:      options.Host,
		apiClient: harborapi.NewAPIClient(cfg),
	}

	return service, nil
}

type ignoreRepositoryService struct {
}

func (i *ignoreRepositoryService) LatestTag(_, _, _ string) (tag string, err error) {
	return "", &NotSupportRegisterError{host: ""}
}

type harborRepositoryService struct {
	host string

	apiClient *harborapi.APIClient
}

func (h *harborRepositoryService) LatestTag(host, projectName, repoName string) (tag string, err error) {
	if host != h.host {
		return tag, &NotSupportRegisterError{h.host}
	}

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
