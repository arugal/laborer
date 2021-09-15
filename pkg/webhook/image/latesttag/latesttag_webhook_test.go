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

import "testing"

func Test_analysisImage(t *testing.T) {
	type args struct {
		image string
	}
	tests := []struct {
		name        string
		args        args
		wantHost    string
		wantProject string
		wantRepo    string
		wantTag     string
		wantPart    int
		wantErr     bool
	}{
		{
			name: "case 1",
			args: args{
				image: "docker.io/pro/repo:1.0.0",
			},
			wantHost:    "docker.io",
			wantProject: "pro",
			wantRepo:    "repo",
			wantTag:     "1.0.0",
			wantPart:    3,
			wantErr:     false,
		},
		{
			name: "case 2",
			args: args{
				image: "pro/repo:1.0.0",
			},
			wantProject: "pro",
			wantRepo:    "repo",
			wantTag:     "1.0.0",
			wantPart:    2,
			wantErr:     false,
		},
		{
			name: "case 3",
			args: args{
				image: "repo:1.0.0",
			},
			wantRepo: "repo",
			wantTag:  "1.0.0",
			wantPart: 1,
			wantErr:  false,
		},
		{
			name: "case 4",
			args: args{
				image: "repo",
			},
			wantRepo: "repo",
			wantTag:  defaultTagName,
			wantPart: 1,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotProject, gotRepo, gotTag, gotPart, err := analysisImage(tt.args.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("analysisImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHost != tt.wantHost {
				t.Errorf("analysisImage() gotHost = %v, want %v", gotHost, tt.wantHost)
			}
			if gotProject != tt.wantProject {
				t.Errorf("analysisImage() gotProject = %v, want %v", gotProject, tt.wantProject)
			}
			if gotRepo != tt.wantRepo {
				t.Errorf("analysisImage() gotRepo = %v, want %v", gotRepo, tt.wantRepo)
			}
			if gotTag != tt.wantTag {
				t.Errorf("analysisImage() gotTag = %v, want %v", gotTag, tt.wantTag)
			}
			if gotPart != tt.wantPart {
				t.Errorf("analysisImage() gotPart = %v, want %v", gotPart, tt.wantPart)
			}
		})
	}
}
