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

package latesttag

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	repositoryservice "github.com/arugal/laborer/pkg/service/repository"
	"gomodules.xyz/jsonpatch/v2"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	defaultTagName = "latest"
)

// latestTagWebHook 创建 Deployment 时将 initContainers 和 containers 的 image
// 设置为镜像仓库中最新的 tag
type latestTagWebHook struct {
	repoService repositoryservice.RepositoryService
	decoder     *admission.Decoder
}

func NewLatestTagWebHook(repoService repositoryservice.RepositoryService) admission.Handler {
	return &latestTagWebHook{
		repoService: repoService,
	}
}

func (l *latestTagWebHook) Handle(ctx context.Context, req admission.Request) admission.Response {
	klog.V(2).Infof("uid: %s, kind: %s, resource: %s, subResource: %s, RequestKind: %s, dryRun: %t",
		req.UID, req.Kind, req.Resource, req.SubResource, req.RequestKind, *req.DryRun)

	deployment := &appsv1.Deployment{}
	err := l.decoder.Decode(req, deployment)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	var patches []jsonpatch.JsonPatchOperation

	for i, initContainer := range deployment.Spec.Template.Spec.InitContainers {
		//initContainer.Image
		host, project, repo, oldTag, part, err := analysisImage(initContainer.Image)
		if err != nil {
			klog.Errorf("analysisImage [%s] err: %v", initContainer.Image, err)
			continue
		}

		tag, err := l.repoService.LatestTag(host, project, repo)
		if err != nil {
			klog.V(2).Infof("%s get latest tag err: %v", initContainer.Image, err)
			continue
		}
		if tag == oldTag {
			continue
		}
		newImage := generateNewImageName(host, project, repo, tag, part)
		klog.Infof("Replace initContainer %s.%s.%s image %s -> %s, dryRun: %t", deployment.Namespace, deployment.Name,
			initContainer.Name, initContainer.Image, newImage, *req.DryRun)

		patches = append(patches, jsonpatch.JsonPatchOperation{
			Operation: "replace",
			Path:      fmt.Sprintf("/spec/template/spec/initContainers/%d/image", i),
			Value:     newImage,
		})
	}

	for i, container := range deployment.Spec.Template.Spec.Containers {
		//container.Image
		host, project, repo, oldTag, part, err := analysisImage(container.Image)
		if err != nil {
			klog.Errorf("analysisImage [%s] err: %v", container.Image, err)
			continue
		}

		tag, err := l.repoService.LatestTag(host, project, repo)
		if err != nil {
			klog.V(2).Infof("%s get latest tag err: %v", container.Image, err)
			continue
		}
		if tag == oldTag {
			continue
		}

		newImage := generateNewImageName(host, project, repo, tag, part)
		klog.Infof("Replace container %s.%s.%s image %s -> %s, dryRun: %t", deployment.Namespace, deployment.Name,
			container.Name, container.Image, newImage, *req.DryRun)

		patches = append(patches, jsonpatch.JsonPatchOperation{
			Operation: "replace",
			Path:      fmt.Sprintf("/spec/template/spec/containers/%d/image", i),
			Value:     newImage,
		})
	}

	resp := admission.Allowed("")
	if len(patches) > 0 {
		resp.Patches = patches
	}
	return resp
}

// InjectDecoder inject the decoder
func (l *latestTagWebHook) InjectDecoder(d *admission.Decoder) error {
	l.decoder = d
	return nil
}

// generateNewImageName 生成新的 image 名称
func generateNewImageName(host, project, repo, tag string, part int) string {
	switch part {
	case 3:
		return fmt.Sprintf("%s/%s/%s:%s", host, project, repo, tag)
	case 2:
		return fmt.Sprintf("%s/%s:%s", project, repo, tag)
	default:
		return fmt.Sprintf("%s:%s", repo, tag)
	}
}

// analysisImage 从 image 中提取信息
func analysisImage(image string) (host, project, repo, tag string, part int, err error) {
	parts := strings.Split(image, "/")
	if len(parts) == 3 {
		// host and project and repo
		host = parts[0]
		project = parts[1]
		repo = parts[2]
	} else if len(parts) == 2 {
		// project and repo
		project = parts[0]
		repo = parts[1]
	} else {
		repo = parts[0]
	}
	repo, tag = analysisTag(repo)
	part = len(parts)
	return
}

// analysisTag 从 repo 中提取信息
func analysisTag(repo string) (name, tag string) {
	parts := strings.Split(repo, ":")
	if len(parts) == 2 {
		name = parts[0]
		tag = parts[1]
	} else {
		name = parts[0]
		tag = defaultTagName
	}
	return
}
