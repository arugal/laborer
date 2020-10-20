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
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

const (
	resourceName = "configmaps"

	configNameSuffix = "-config"

	restartedAt = "kubectl.kubernetes.io/restartedAt"
)

func init() {
	namespace.RegisterNewControllerFunc(newConfigmapControllerFunc())
}

type configmapController struct {
	namespace.BaseController

	stopCh   chan struct{}
	indexer  cache.Indexer
	informer cache.Controller
}

// newConfigmapControllerFunc
func newConfigmapControllerFunc() namespace.NewControllerFunc {
	return func(ns string, k8sClient kubernetes.Interface, informers informers.InformerFactory) namespace.Controller {
		configMapListWatcher := cache.NewListWatchFromClient(k8sClient.CoreV1().RESTClient(), resourceName, ns, fields.Everything())

		deploymentsClient := k8sClient.AppsV1().Deployments(ns)

		indexer, informer := cache.NewIndexerInformer(configMapListWatcher, &v1.ConfigMap{}, 0, cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				// ignore
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				configMap := newObj.(*v1.ConfigMap)

				if !strings.HasSuffix(configMap.Name, configNameSuffix) {
					return
				}

				deploymentName := strings.TrimSuffix(configMap.Name, configNameSuffix)
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
					klog.Errorf("ConfigMap %s controller marshal %v err: %s", ns, newDeployment, err)
					return
				}
				if klog.V(2) {
					klog.Infof("ConfigMap %s controller restart %s deployment: %s", ns, deploymentName, string(data))
				}

				_, err = deploymentsClient.Patch(deploymentName, types.StrategicMergePatchType, data)
				if err != nil {
					klog.Errorf("ConfigMap %s controller patch %v err: %s", ns, string(data), err)
				}
			},
			DeleteFunc: func(obj interface{}) {
				// ignore
			},
		}, cache.Indexers{})
		return &configmapController{
			BaseController: namespace.BaseController{
				NameSpace: ns,
			},
			stopCh:   make(chan struct{}),
			indexer:  indexer,
			informer: informer,
		}
	}
}

func (c *configmapController) Run() {
	defer runtime.HandleCrash()

	klog.Infof("Start ConfigMap controller from namespace: %s", c.NameSpace)

	go c.informer.Run(c.stopCh)

	if !cache.WaitForCacheSync(c.stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("%s Timed out waiting for caches to sync", c.NameSpace))
		return
	}
}

func (c *configmapController) Stop() {
	klog.Infof("Stopping ConfigMap controller from namespace: %s ", c.NameSpace)
	close(c.stopCh)
}
