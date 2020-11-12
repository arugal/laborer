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

package event

import (
	"fmt"
	"strings"

	"github.com/arugal/laborer/pkg/crash"
)

type ImageEvent struct {
	Image string `json:"image"`
	Tag   string `json:"tag"`
}

func (e ImageEvent) String() string {
	return fmt.Sprintf("%s:%s", e.Image, e.Tag)
}

func OfImageEvent(image string) ImageEvent {
	imageEvent := ImageEvent{
		Image: image,
		Tag:   "latest",
	}
	if strings.Contains(image, ":") {
		index := strings.LastIndex(image, ":")
		imageEvent.Image = image[:index]
		imageEvent.Tag = image[index+1:]
	}
	return imageEvent
}

// ImageEventHandlerFunc 处理镜像时间的函数
type ImageEventHandlerFunc func(event ImageEvent)

// ImageEventCollect 收集镜像中心的 webhook(harbor) 事件, 并回调 ImageEventHandlerFunc
type ImageEventCollect interface {
	// 收集 webhook 产生的事件
	Collect(event ImageEvent)
	// 注册事件处理函数
	RegisterHandlerFunc(handler ImageEventHandlerFunc)

	Start(stop <-chan struct{})
}

// NewImageEventCollect
func NewImageEventCollect() ImageEventCollect {
	return &defaultImageEventCollect{
		eventCh: make(chan ImageEvent, 100),
	}
}

// defaultImageEventCollect 收集器默认实现
type defaultImageEventCollect struct {
	eventCh      chan ImageEvent
	handlerFuncs []ImageEventHandlerFunc
}

func (d *defaultImageEventCollect) Collect(event ImageEvent) {
	d.eventCh <- event
}

func (d *defaultImageEventCollect) RegisterHandlerFunc(handler ImageEventHandlerFunc) {
	d.handlerFuncs = append(d.handlerFuncs, handler)
}

func (d *defaultImageEventCollect) Start(stop <-chan struct{}) {
	warpHandlerFunc := func(event ImageEvent, handlerFunc ImageEventHandlerFunc) {
		defer crash.HandleCrash(crash.DefaultHandler)
		handlerFunc(event)
	}

	go func() {
		defer close(d.eventCh)
		for {
			select {
			case <-stop:
				return
			case event := <-d.eventCh:
				for _, f := range d.handlerFuncs {
					warpHandlerFunc(event, f)
				}
			}
		}
	}()
}
