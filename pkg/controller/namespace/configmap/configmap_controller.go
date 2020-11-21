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

package configmap

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	k8sv1 "github.com/arugal/laborer/pkg/api/k8s/v1"
	"github.com/arugal/laborer/pkg/controller/namespace"
	"github.com/arugal/laborer/pkg/informers"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

const (
	configNameSuffix = "-config"

	restartedAt = "kubectl.kubernetes.io/restartedAt"

	annotationName = "laborer.configmap.associate.deployment"
)

func init() {
	namespace.RegisterNewControllerFunc(newConfigmapControllerFunc())
}

// configmapController 当 configmap 变化时重新部署对应的 deployment
type configmapController struct {
	namespace.BaseController

	stopCh                  chan struct{}
	configmapInformerSynced cache.InformerSynced
}

// newConfigmapControllerFunc
func newConfigmapControllerFunc() namespace.NewControllerFunc {
	return func(ns string, k8sClient kubernetes.Interface, namespaceInformerFactory informers.InformerFactory) namespace.Controller {
		deploymentsLister := namespaceInformerFactory.KubernetesSharedInformerFactory().Apps().V1().Deployments().Lister()
		deploymentsClient := k8sClient.AppsV1().Deployments(ns)

		handlerFunc := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if klog.V(2) {
					configmap := obj.(*v1.ConfigMap)
					klog.Infof("configmap add: %s.%s", ns, configmap.Name)
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				newConfigmap := newObj.(*v1.ConfigMap)
				oldConfigmap := oldObj.(*v1.ConfigMap)
				if oldConfigmap.ResourceVersion == newConfigmap.ResourceVersion {
					klog.V(2).Infof("configmap update %s.%s, oldVersion: %s, newVersion: %s, ignored",
						newConfigmap.Namespace, newConfigmap.Name, newConfigmap.ResourceVersion, oldConfigmap.ResourceVersion)
					// ignored (by resync)
					return
				}

				klog.V(2).Infof("configmap update: %s.%s", ns, newConfigmap.Name)
				needRestartDeployments := analyzeDeployments(newConfigmap)

				for _, deploymentName := range needRestartDeployments {
					deploy, err := deploymentsLister.Deployments(ns).Get(deploymentName)
					if err != nil || deploy == nil {
						if !errors.IsNotFound(err) {
							klog.Errorf("[%s] get deployment [%s] err: %v", ns, deploymentName, err)
						}
						continue
					}

					newDeployment := k8sv1.Deployment{
						Spec: k8sv1.DeploymentSpec{
							Template: k8sv1.PodTemplateSpec{
								Metadata: k8sv1.Metadata{
									Annotations: map[string]string{
										restartedAt: time.Now().Format(time.RFC3339),
									},
								},
							},
						},
					}

					data, err := json.Marshal(newDeployment)
					if err != nil {
						klog.Errorf("configmap [%s] controller marshal %v err: %s", ns, newDeployment, err)
						return
					}

					klog.Infof("configmap trigger %s.%s restarted", newConfigmap.Namespace, deploymentName)
					if _, err = deploymentsClient.Patch(deploymentName, types.StrategicMergePatchType, data); err != nil {
						klog.Errorf("configmap [%s] controller patch %v err: %s", ns, string(data), err)
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				if klog.V(2) {
					configmap := obj.(*v1.ConfigMap)
					klog.Infof("configmap delete: %s.%s", ns, configmap.Name)
				}
			},
		}

		informer := namespaceInformerFactory.KubernetesSharedInformerFactory().Core().V1().ConfigMaps().Informer()
		informer.AddEventHandler(handlerFunc)

		return &configmapController{
			BaseController: namespace.BaseController{
				NameSpace: ns,
			},
			stopCh:                  make(chan struct{}),
			configmapInformerSynced: informer.HasSynced,
		}
	}
}

func (c *configmapController) Run() {
	defer runtime.HandleCrash()
	klog.Infof("Starting configmap controller from namespace: %s", c.NameSpace)

	if !cache.WaitForCacheSync(c.stopCh, c.configmapInformerSynced) {
		runtime.HandleError(fmt.Errorf("%s Timed out waiting for caches to sync", c.NameSpace))
		return
	}
}

func (c *configmapController) Stop() {
	klog.Infof("Stopping configmap controller from namespace: %s ", c.NameSpace)
	close(c.stopCh)
}

// analyzeDeployments analyze configmap associated with deployment
func analyzeDeployments(configmap *v1.ConfigMap) (deployments []string) {
	// 通过 map 过滤重复的 deployment name
	var deploymentsMap = make(map[string]bool)

	// step1. 根据 configmap 的名称解析 deployment 的名称
	if strings.HasSuffix(configmap.Name, configNameSuffix) {
		deploymentsMap[strings.TrimSuffix(configmap.Name, configNameSuffix)] = true
	}

	// step2. 从 annotation 中提取 configmap 关联的 deployment 的名称
	if annotation, ok := configmap.Annotations[annotationName]; ok {
		var deploys []string
		err := json.Unmarshal([]byte(annotation), &deploys)
		if err != nil {
			klog.Errorf("Unmarshal configmap [%s.%s] annotation [%s] err: %v", configmap.Namespace, configmap.Name, annotation, err)
			return
		}
		for _, deploy := range deploys {
			deploymentsMap[deploy] = true
		}
	}
	// map to since
	for k := range deploymentsMap {
		deployments = append(deployments, k)
	}
	return
}
