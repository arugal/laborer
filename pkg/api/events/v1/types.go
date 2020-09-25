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

package v1

import (
	"fmt"
	"strings"

	"github.com/arugal/laborer/pkg/crash"
)

type ImageEvent struct {
	Image string `json:"image"`
	Tag   string `json:"tag"`
}

func (e ImageEvent) ImageAndTag() string {
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

// ImageEventHandlerFunc
type ImageEventHandlerFunc func(event ImageEvent)

// ImageEventInterface collect and process image event
type ImageEventInterface interface {
	// collect new image event
	Collect(event ImageEvent)
	// register handler func
	AddImageEventFunc(handler ImageEventHandlerFunc)

	Start(stop <-chan struct{})
}

func NewImageEventInterface() ImageEventInterface {
	return &defaultImageEventInterface{
		eventCh: make(chan ImageEvent, 100),
	}
}

type defaultImageEventInterface struct {
	eventCh      chan ImageEvent
	handlerFuncs []ImageEventHandlerFunc
}

func (d *defaultImageEventInterface) Collect(event ImageEvent) {
	d.eventCh <- event
}

func (d *defaultImageEventInterface) AddImageEventFunc(handler ImageEventHandlerFunc) {
	d.handlerFuncs = append(d.handlerFuncs, handler)
}

func (d *defaultImageEventInterface) Start(stop <-chan struct{}) {
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
