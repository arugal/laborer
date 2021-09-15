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

package v1

type Deployment struct {
	Spec DeploymentSpec `json:"spec"`
}

type DeploymentSpec struct {
	Template PodTemplateSpec `json:"template"`
}

type PodTemplateSpec struct {
	Metadata Metadata `json:"metadata,omitempty"`
	Spec     PodSpec  `json:"spec,omitempty"`
}

type Metadata struct {
	Annotations map[string]string `json:"annotations"`
}

type PodSpec struct {
	Containers []Container `json:"containers,omitempty"`
}

type Container struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}
