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
	"sort"

	"github.com/scultura-org/harborapi"
)

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
