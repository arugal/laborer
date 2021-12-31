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

package job

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/arugal/laborer/pkg/controller/namespace"
	"github.com/arugal/laborer/pkg/crash"
	"github.com/arugal/laborer/pkg/informers"
	eventservice "github.com/arugal/laborer/pkg/service/event"
	apibatchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	v1 "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

const (
	rerunAnnotation = "laborer.job.rerun"
	disableRerun    = "false"

	retryTimes = 3
)

type jobController struct {
	namespace.BaseController

	stopCh chan struct{}

	jobInformerSynced cache.InformerSynced
	jobLister         v1.JobLister

	jobClient batchv1.JobInterface
}

func newJobControllerFunc(ns string, k8sClient kubernetes.Interface, namespaceInformerFactory informers.InformerFactory) namespace.Controller {
	jobLister := namespaceInformerFactory.KubernetesSharedInformerFactory().Batch().V1().Jobs().Lister()
	jobInformer := namespaceInformerFactory.KubernetesSharedInformerFactory().Batch().V1().Jobs().Informer()

	jobsClient := k8sClient.BatchV1().Jobs(ns)

	return &jobController{
		BaseController: namespace.BaseController{
			NameSpace: ns,
		},
		stopCh:            make(chan struct{}),
		jobLister:         jobLister,
		jobInformerSynced: jobInformer.HasSynced,
		jobClient:         jobsClient,
	}
}

func (j *jobController) Run() {
	defer crash.HandleCrash()
	klog.Infof("Starting job controller from namespace: %s", j.NameSpace)

	if !cache.WaitForCacheSync(j.stopCh, j.jobInformerSynced) {
		runtime.HandleError(fmt.Errorf("%s Timed out waiting for caches to sync", j.NameSpace))
		return
	}
}

func (j *jobController) Stop() {
	klog.Infof("Stopping job controller for namespace: %s", j.NameSpace)
	close(j.stopCh)
}

func (j *jobController) ProcessImageEvent(event eventservice.ImageEvent) {
	defer crash.HandleCrash(crash.DefaultHandler)

	jobs, err := j.jobLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("[%s] list job err: %v", j.NameSpace, err)
		return
	}

	for _, job := range jobs {
		if value, ok := job.Annotations[rerunAnnotation]; !ok || value == disableRerun {
			// 只有当设置了 laborer.job.rerun 的 job 才会被重新执行
			continue
		}

		var rerun bool
		for _, container := range job.Spec.Template.Spec.Containers {
			containerImage := eventservice.OfImageEvent(container.Image)
			if containerImage.Image == event.Image {
				if containerImage.Tag != event.Tag || container.ImagePullPolicy == corev1.PullAlways {
					rerun = true
					containerImage.Tag = event.Tag
				}
			}
		}
		if !rerun {
			continue
		}
		j.jobReRun(job)
	}
}

func (j *jobController) jobReRun(job *apibatchv1.Job) {
	newJob := *job
	newJob.ResourceVersion = ""
	newJob.Status = apibatchv1.JobStatus{}
	newJob.ObjectMeta.UID = ""
	newJob.Annotations["revisions"] = strings.Replace(job.Annotations["revisions"], "running", "unfinished", -1)

	delete(newJob.Spec.Selector.MatchLabels, "controller-uid")
	delete(newJob.Spec.Template.ObjectMeta.Labels, "controller-uid")

	err := j.jobClient.Delete(context.Background(), job.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("failed to rerun job %s/%s, reason: %s", j.NameSpace, job.Name, err)
		return
	}

	for i := 0; i < retryTimes; i++ {
		_, err = j.jobClient.Create(context.Background(), &newJob, metav1.CreateOptions{})
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		return
	}
	if err != nil {
		klog.Errorf("failed to rerun job %s/%s, reason: %s", j.NameSpace, job.Name, err)
	}
}
