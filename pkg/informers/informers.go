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

package informers

import (
	"time"

	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

// default re-sync period for all informer factories
const defaultResync = 600 * time.Second

type InformerFactory interface {
	KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory

	// Start shared informer factory one by one if they are not nil
	Start(stopCh <-chan struct{})
}

type informerFactories struct {
	informerFactory k8sinformers.SharedInformerFactory
}

// NewInformerFactories
func NewInformerFactories(client kubernetes.Interface) InformerFactory {
	factory := &informerFactories{}

	if client != nil {
		factory.informerFactory = k8sinformers.NewSharedInformerFactory(client, defaultResync)
	}

	return factory
}

func (f *informerFactories) KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory {
	return f.informerFactory
}

func (f *informerFactories) Start(stop <-chan struct{}) {
	if f.informerFactory != nil {
		f.informerFactory.Start(stop)
	}
}
