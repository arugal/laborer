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

package harbor

const (
	Push = "PUSH_ARTIFACT"
)

//{
//		"type":"PUSH_ARTIFACT",
//	 	"occur_at":1603728502,
//	 	"operator":"admin",
//	 	"event_data":{
//	 		"resources":[
//	 			{
//	 				"digest":"sha256:65fffb1482321b23ed3fc24bd6961385335ec7fca12de3420a9d778afe3c5e56",
//	 				"tag":"v1.0.0","resource_url":"image/image:v1.0.0"
//	 			}
//			],
//			"repository":
//			{
//				"date_created":1600856703,
//				"name":"image",
//				"namespace":"image",
//				"repo_full_name":"image/image",
//				"repo_type":"private"
//			}
//		}
//	}

type WebHook struct {
	Type      string    `json:"type"`
	OccurAt   int32     `json:"occur_at,omitempty"`
	Operator  string    `json:"operator,omitempty"`
	EventData EventData `json:"event_data"`
}

type EventData struct {
	Resources  []EventResource `json:"resources"`
	Repository Repository      `json:"repository"`
}

type EventResource struct {
	Digest      string `json:"digest,omitempty"`
	Tag         string `json:"tag"`
	ResourceURL string `json:"resource_url"`
}

type Repository struct {
	DateCreated  int32  `json:"date_created,omitempty"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	RepoFullName string `json:"repo_full_name"`
	RepoType     string `json:"repo_type"`
}
