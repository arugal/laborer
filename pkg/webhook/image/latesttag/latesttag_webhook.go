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

package latesttag

import (
	"context"
	"fmt"
	"net/http"

	repositoryservice "github.com/arugal/laborer/pkg/service/repository"
	"gomodules.xyz/jsonpatch/v2"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
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
	deployment := &appsv1.Deployment{}
	err := l.decoder.Decode(req, deployment)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	var patches []jsonpatch.JsonPatchOperation

	for i, initContainer := range deployment.Spec.Template.Spec.InitContainers {
		//initContainer.Image

		patches = append(patches, jsonpatch.JsonPatchOperation{
			Operation: "merge",
			Path:      fmt.Sprintf("spec.template.spec.initContainers[%d].image", i),
			Value:     initContainer.Image,
		})
	}

	for i, container := range deployment.Spec.Template.Spec.Containers {
		//container.Image
		patches = append(patches, jsonpatch.JsonPatchOperation{
			Operation: "merge",
			Path:      fmt.Sprintf("spec.template.spec.containers[%d].image", i),
			Value:     container.Image,
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
