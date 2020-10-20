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

package deployment

import (
	"encoding/json"
	"fmt"

	eventsv1 "github.com/arugal/laborer/pkg/api/events/v1"
	k8sv1 "github.com/arugal/laborer/pkg/api/k8s/v1"
	"github.com/arugal/laborer/pkg/controller/namespace"
	"github.com/arugal/laborer/pkg/crash"
	"github.com/arugal/laborer/pkg/informers"
	apiappsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

func init() {
	namespace.RegisterNewControllerFunc(newDeploymentControllerFunc)
}

const resourceName = "deployments"

type deploymentController struct {
	namespace.BaseController

	stopCh   chan struct{}
	indexer  cache.Indexer
	informer cache.Controller

	deploymentsClient appsv1.DeploymentInterface
}

// newDeploymentControllerFunc
func newDeploymentControllerFunc(ns string, k8sClient kubernetes.Interface, informers informers.InformerFactory) namespace.Controller {
	deploymentListWatcher := cache.NewListWatchFromClient(k8sClient.AppsV1().RESTClient(), resourceName, ns, fields.Everything())

	indexer, informer := cache.NewIndexerInformer(deploymentListWatcher, &apiappsv1.Deployment{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			klog.V(2).Infof("Deployment %s controller add: %v", ns, obj)
		},
		UpdateFunc: func(old, new interface{}) {
			klog.V(2).Infof("Deployment %s controller add: %v", ns, new)
		},
		DeleteFunc: func(obj interface{}) {
			klog.V(2).Infof("Deployment %s controller add: %v", ns, obj)
		},
	}, cache.Indexers{})

	deploymentsClient := k8sClient.AppsV1().Deployments(ns)

	return &deploymentController{
		BaseController: namespace.BaseController{
			NameSpace: ns,
		},
		stopCh:            make(chan struct{}),
		indexer:           indexer,
		informer:          informer,
		deploymentsClient: deploymentsClient,
	}
}

func (d *deploymentController) Run() {
	defer crash.HandleCrash()
	klog.Infof("Start deployment controller from namespace: %s", d.NameSpace)

	go d.informer.Run(d.stopCh)

	if !cache.WaitForCacheSync(d.stopCh, d.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("%s Timed out waiting for caches to sync", d.NameSpace))
		return
	}

	if klog.V(2) {
		for _, deployment := range d.indexer.List() {
			klog.Infof("Deployment controller %v from namespace: %s", deployment, d.NameSpace)
		}
	}
}

func (d *deploymentController) Stop() {
	klog.Infof("Stopping Deployment controller from namespace: %s", d.NameSpace)
	close(d.stopCh)
}

func (d *deploymentController) ProcessImageEvent(event eventsv1.ImageEvent) {
	defer crash.HandleCrash(crash.DefaultHandler)

	for _, object := range d.indexer.List() {
		deployment := object.(*apiappsv1.Deployment)
		var updateContainers []k8sv1.Container

		for _, container := range deployment.Spec.Template.Spec.Containers {
			containerImage := eventsv1.OfImageEvent(container.Image)
			if containerImage.Image == event.Image && containerImage.Tag != event.Tag {
				updateContainers = append(updateContainers, k8sv1.Container{
					Name:  container.Name,
					Image: event.ImageAndTag(),
				})
			}
		}

		if len(updateContainers) > 0 {
			newDeployment := k8sv1.Deployment{
				Spec: k8sv1.DeploymentSpec{
					Template: k8sv1.PodTemplateSpec{
						Spec: k8sv1.PodSpec{
							Containers: updateContainers,
						},
					},
				},
			}

			data, err := json.Marshal(newDeployment)
			if err != nil {
				klog.Error(err)
				return
			}

			if klog.V(2) {
				klog.Infof("Deployment %s controller update container image: %s", d.NameSpace, string(data))
			}

			_, err = d.deploymentsClient.Patch(deployment.Name, types.StrategicMergePatchType, data)
		}
	}
}
