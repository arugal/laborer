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

package harbor

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	eventsv1 "github.com/arugal/laborer/pkg/api/events/v1"
	"k8s.io/klog"
)

type imageEventWebHook struct {
	collect eventsv1.ImageEventCollect
}

func NewImageEventWebHook(collect eventsv1.ImageEventCollect) http.Handler {
	return &imageEventWebHook{
		collect: collect,
	}
}

func (i *imageEventWebHook) ServeHTTP(_ http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		klog.Warningf("Read harbor webhook body error: %v", err)
		return
	}
	if len(body) == 0 {
		klog.Info("Harbor webhook body is empty")
		return
	}
	var bodyMap map[string]interface{}
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		klog.Warningf("Unmarshal harbor webhook body [%s] error: %v", string(body), err)
		return
	}
	if typ, ok := bodyMap["type"]; !ok || typ != Push {
		klog.Infof("Unsupported event type %s, ignored", string(body))
		return
	}

	var webhook WebHook
	err = json.Unmarshal(body, &webhook)
	if err != nil {
		klog.Warningf("Unmarshal harbor webhook body [%s] error: %v", string(body), err)
		return
	}

	if klog.V(4) {
		klog.Info("Harbor event data: %s", string(body))
	}

	for _, resource := range webhook.EventData.Resources {
		i.collect.Collect(eventsv1.OfImageEvent(resource.ResourceURL))
	}
}
