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

import "testing"

func Test_mockRepositoryService_LatestTag(t *testing.T) {
	type fields struct {
		tags map[string]string
	}
	type args struct {
		host        string
		projectName string
		repoName    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantTag string
		wantErr bool
	}{
		{
			name: "normal1",
			fields: fields{
				tags: map[string]string{
					"image": "v2",
				},
			},
			args: args{
				host:        "",
				projectName: "",
				repoName:    "image",
			},
			wantTag: "v2",
			wantErr: false,
		},
		{
			name: "normal2",
			fields: fields{
				tags: map[string]string{
					"project/image": "v2",
				},
			},
			args: args{
				host:        "",
				projectName: "project",
				repoName:    "image",
			},
			wantTag: "v2",
			wantErr: false,
		},
		{
			name: "normal3",
			fields: fields{
				tags: map[string]string{
					"register/project/image": "v2",
				},
			},
			args: args{
				host:        "register",
				projectName: "project",
				repoName:    "image",
			},
			wantTag: "v2",
			wantErr: false,
		},
		{
			name: "not found",
			fields: fields{
				tags: map[string]string{},
			},
			args: args{
				host:        "register",
				projectName: "project",
				repoName:    "image",
			},
			wantTag: "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mockRepositoryService{
				tags: tt.fields.tags,
			}
			gotTag, err := m.LatestTag(tt.args.host, tt.args.projectName, tt.args.repoName)
			if (err != nil) != tt.wantErr {
				t.Errorf("LatestTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTag != tt.wantTag {
				t.Errorf("LatestTag() gotTag = %v, want %v", gotTag, tt.wantTag)
			}
		})
	}
}
