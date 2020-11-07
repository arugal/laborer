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
	ht "net/http"
	"strings"

	"github.com/antihax/optional"
	"github.com/scultura-org/harborapi"
	"k8s.io/klog"
)

type protocol string

const (
	http  protocol = "http"
	https protocol = "https"
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

// RepositoryServiceOption 设置 repository service
type RepositoryServiceOption func(service *harborRepositoryService)

// WithHttp default use https
func WithHttp() RepositoryServiceOption {
	return func(service *harborRepositoryService) {
		service.protocol = http
	}
}

// WithInsecureSkipVerify
func WithInsecureSkipVerify(insecureSkipVerify bool) RepositoryServiceOption {
	return func(service *harborRepositoryService) {
		service.insecureSkipVerify = insecureSkipVerify
	}
}

// WithPathPrefix default is /api/v2.0
func WithPathPrefix(pathPrefix string) RepositoryServiceOption {
	return func(service *harborRepositoryService) {
		if !strings.HasPrefix(pathPrefix, "/") {
			pathPrefix = "/" + pathPrefix
		}
		service.pathPrefix = pathPrefix
	}
}

// WithHost
func WithHost(host string) RepositoryServiceOption {
	return func(service *harborRepositoryService) {
		service.host = host
	}
}

// TODO support multiple register
// NewRepositoryService
func NewRepositoryService(opts ...RepositoryServiceOption) (RepositoryService, error) {
	service := &harborRepositoryService{
		protocol:   https,
		pathPrefix: "/api/v2.0",
	}

	// with options
	for _, opt := range opts {
		opt(service)
	}

	if service.host == "" {
		return nil, errors.New("host must be set")
	}

	if service.pathPrefix == "" {
		klog.V(2).Infof("pathPrefix is empty")
	}

	cfg := harborapi.NewConfiguration()
	cfg.BasePath = fmt.Sprintf("%s://%s%s", service.protocol, service.host, service.pathPrefix)

	// insecureSkipVerify
	if service.insecureSkipVerify {
		tr := &ht.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		cfg.HTTPClient = &ht.Client{
			Transport: tr,
		}
	}

	service.apiClient = harborapi.NewAPIClient(cfg)

	return service, nil
}

type harborRepositoryService struct {
	protocol   protocol
	host       string
	pathPrefix string

	insecureSkipVerify bool

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
