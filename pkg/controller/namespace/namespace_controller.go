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
	"fmt"

	"github.com/arugal/laborer/pkg/crash"
	"github.com/arugal/laborer/pkg/informers"
	eventservice "github.com/arugal/laborer/pkg/service/event"
	v1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	informerv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

type eventType int

const (
	added    eventType = iota
	modified eventType = iota
	deleted  eventType = iota

	laborerEnable = "laborer.enable"
	enabled       = "true"
)

// +kubebuilder:rbac:groups="",resources=configmaps;namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=list;watch;patch

// NamespaceController namespace 控制器，根据 namespace 的 labels 判断是否启动 AggregationController
type NamespaceController struct {
	client kubernetes.Interface

	namespaceInformer       informerv1.NamespaceInformer
	namespaceInformerSynced cache.InformerSynced

	aggregationControllerMap map[string]Controller
}

func NewNamespaceController(informers informers.InformerFactory, client kubernetes.Interface, imageEventCollect eventservice.ImageEventCollect) *NamespaceController {
	n := &NamespaceController{
		client:                   client,
		aggregationControllerMap: map[string]Controller{},
	}

	namespaceInformer := informers.KubernetesSharedInformerFactory().Core().V1().Namespaces()
	namespaceInformer.Informer().AddEventHandler(n.newResourceEventHandlerFuncs())

	n.namespaceInformerSynced = namespaceInformer.Informer().HasSynced

	imageEventCollect.RegisterHandlerFunc(n.ImageEventHandlerFunc)
	return n
}

func (n *NamespaceController) Start(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()

	klog.Info("Starting namespace controller")
	defer klog.Info("shutting down namespace controller")

	if !cache.WaitForCacheSync(stopCh, n.namespaceInformerSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	<-stopCh
	return nil
}

// syncAggregationController
func (n *NamespaceController) syncAggregationController(et eventType, obj interface{}) {
	defer crash.HandleCrash(crash.DefaultHandler)
	namespace := obj.(*v1.Namespace)

	switch et {
	case added:
		if enable, ok := namespace.Labels[laborerEnable]; ok && enable == enabled {
			n.addNewAggregationController(namespace.Name)
		}
	case modified:
		label := namespace.Labels[laborerEnable]
		controller, ok := n.aggregationControllerMap[namespace.Name]

		if label == enabled && !ok {
			n.addNewAggregationController(namespace.Name)
		} else if label != enabled && ok {
			n.stopAggregationController(controller)
		}
	case deleted:
		if controller, ok := n.aggregationControllerMap[namespace.Name]; ok {
			n.stopAggregationController(controller)
		}
	}
}

func (n *NamespaceController) addNewAggregationController(namespace string) {
	controller := NewAggregationController(namespace, n.client)
	controller.Run()
	n.aggregationControllerMap[namespace] = controller
}

func (n *NamespaceController) stopAggregationController(c Controller) {
	c.Stop()
	delete(n.aggregationControllerMap, c.Namespace())
}

// ImageEventHandlerFunc update the deployment container image based on event.
func (n *NamespaceController) ImageEventHandlerFunc(event eventservice.ImageEvent) {
	for _, ctrl := range n.aggregationControllerMap {
		ctrl.ProcessImageEvent(event)
	}
}

func (n *NamespaceController) newResourceEventHandlerFuncs() cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			n.syncAggregationController(added, obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			n.syncAggregationController(modified, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			n.syncAggregationController(deleted, obj)
		},
	}
}
