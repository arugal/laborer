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

package github

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	eventservice "github.com/arugal/laborer/pkg/service/event"
	"k8s.io/klog"
)

const (
	published = "published"

	packageType = "CONTAINER"
)

// imageEventWebhook github webhook
type imageEventWebhook struct {
	collect eventservice.ImageEventCollect
}

func NewImageEventWebhook(collect eventservice.ImageEventCollect) http.Handler {
	return &imageEventWebhook{
		collect: collect,
	}
}

func (i *imageEventWebhook) ServeHTTP(_ http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		klog.Warningf("Read github webhook body error: %v", err)
		return
	}

	if len(body) == 0 {
		klog.Warningf("Github webhook body is empty")
		return
	}

	var bodyMap map[string]interface{}
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		klog.Warningf("Unmarshal github webhook body [%s] error: %v", string(body), err)
		return
	}

	if action, ok := bodyMap["action"]; !ok || action != published {
		klog.Warningf("Unsupported event type %s, ignored", string(body))
		return
	}

	pkg, ok := bodyMap["package"]
	if !ok {
		klog.Warningf("Package not obtained, ignored %+v", pkg)
		return
	}

	pkgByte, err := json.Marshal(pkg)
	if err != nil {
		return
	}

	var pkage Package
	err = json.Unmarshal(pkgByte, &pkage)
	if err != nil {
		klog.Warningf("Unmarshal github webhook package error: %v", err)
		return
	}

	if pkage.PackageType != packageType || pkage.PackageVersion.PackageUrl == "" {
		klog.Warningf("Lack of essential content, ignored %+v", pkage)
		return
	}

	i.collect.Collect(eventservice.OfImageEvent(pkage.PackageVersion.PackageUrl))
}
