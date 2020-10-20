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

package namespace

import (
	eventsv1 "github.com/arugal/laborer/pkg/api/events/v1"
	"github.com/arugal/laborer/pkg/informers"
	"k8s.io/client-go/kubernetes"
)

var (
	newControllerFuncs []NewControllerFunc
)

// RegisterNewControllerFunc
func RegisterNewControllerFunc(newFunc NewControllerFunc) {
	newControllerFuncs = append(newControllerFuncs, newFunc)
}

// Controller Listens for resources in the current namespace
type Controller interface {
	// return the namespace of the current controller
	Namespace() string
	Run()
	Stop()
	ProcessImageEvent(event eventsv1.ImageEvent)
}

type NewControllerFunc func(namespace string, k8sClient kubernetes.Interface, informers informers.InformerFactory) Controller

// BaseController empty implementation
type BaseController struct {
	NameSpace string
}

func (b BaseController) Namespace() string {
	return b.NameSpace
}

func (b BaseController) ProcessImageEvent(eventsv1.ImageEvent) {

}

// aggregationController aggregate multiple controller, such as deployments, configmap.
type aggregationController struct {
	BaseController

	k8sClient kubernetes.Interface
	informers informers.InformerFactory

	controllers []Controller
}

func NewAggregationController(namespace string, k8sClient kubernetes.Interface, informers informers.InformerFactory) Controller {
	c := &aggregationController{
		BaseController: BaseController{
			NameSpace: namespace,
		},
		k8sClient: k8sClient,
		informers: informers,
	}

	for _, newFunc := range newControllerFuncs {
		c.controllers = append(c.controllers, newFunc(namespace, k8sClient, informers))
	}

	return c
}

func (a *aggregationController) Run() {
	for _, c := range a.controllers {
		c.Run()
	}
}

func (a *aggregationController) Stop() {
	for _, c := range a.controllers {
		c.Stop()
	}
}

func (a *aggregationController) ProcessImageEvent(event eventsv1.ImageEvent) {
	for _, c := range a.controllers {
		c.ProcessImageEvent(event)
	}
}