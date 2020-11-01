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

package app

import (
	"fmt"
	"os"

	"github.com/arugal/laborer/cmd/controller-manager/app/options"
	"github.com/arugal/laborer/pkg/config"
	"github.com/arugal/laborer/pkg/controller/namespace"
	"github.com/arugal/laborer/pkg/informers"
	eventservice "github.com/arugal/laborer/pkg/service/event"
	repositoryservice "github.com/arugal/laborer/pkg/service/repository"
	"github.com/arugal/laborer/pkg/simple/client/k8s"
	"github.com/arugal/laborer/pkg/utils/term"
	"github.com/arugal/laborer/pkg/webhook/image/harbor"
	"github.com/arugal/laborer/pkg/webhook/image/latesttag"
	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func NewControllerManagerCommand() *cobra.Command {
	s := options.NewLaborerControllerManagerOptions()
	conf, err := config.TryLoadFromDisk()
	if err != nil {
		klog.Warning("Failed to load configuration from disk", err)
	} else {
		s = &options.LaborerControllerManagerOptions{
			KubernetesOptions:    conf.KubernetesOptions,
			LeaderElection:       s.LeaderElection,
			LeaderElectNamespace: s.LeaderElectNamespace,
			LeaderElect:          s.LeaderElect,
			WebhookCertDir:       s.WebhookCertDir,
		}
	}

	cmd := &cobra.Command{
		Use:  "controller-manager",
		Long: "Laborer controller manager is  a daemon that",
		Run: func(cmd *cobra.Command, args []string) {
			if errs := s.Validate(); len(errs) > 0 {
				klog.Error(utilerrors.NewAggregate(errs))
				os.Exit(1)
			}

			if err := run(s, signals.SetupSignalHandler()); err != nil {
				klog.Error(err)
				os.Exit(1)
			}
		},
		SilenceUsage: true,
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})
	return cmd
}

func run(s *options.LaborerControllerManagerOptions, stopCh <-chan struct{}) error {
	kubernetesClient, err := k8s.NewKubernetesClient(s.KubernetesOptions)
	if err != nil {
		klog.Errorf("Failed to create kubernetes clientset %v", err)
		return err
	}

	informerFactory := informers.NewInformerFactories(kubernetesClient.Kubernetes())

	mgrOptions := manager.Options{
		CertDir: s.WebhookCertDir,
		Port:    8443,
	}

	if s.LeaderElect {
		mgrOptions = manager.Options{
			CertDir:                 s.WebhookCertDir,
			Port:                    8443,
			LeaderElection:          s.LeaderElect,
			LeaderElectionNamespace: s.LeaderElectNamespace,
			LeaderElectionID:        "laborer-controller-manager-leader-election",
			LeaseDuration:           &s.LeaderElection.LeaseDuration,
			RetryPeriod:             &s.LeaderElection.RetryPeriod,
			RenewDeadline:           &s.LeaderElection.RenewDeadline,
		}
	}

	klog.V(0).Info("setting up manager")

	imageEventCollect := eventservice.NewImageEventCollect()
	repositoryService, err := repositoryservice.NewRepositoryService()
	if err != nil {
		klog.Fatalf("NewRepositoryService err: %v\n", err)
	}

	// Use 8443 instead of 443 cause we need root permission to bind port 443
	mgr, err := manager.New(kubernetesClient.Config(), mgrOptions)
	if err != nil {
		klog.Fatalf("unable to set up overall controller manager: %v", err)
	}

	namespaceController := namespace.NewNamespaceController(informerFactory, kubernetesClient.Kubernetes(), imageEventCollect)

	controllers := map[string]manager.Runnable{
		"namespace-controller": namespaceController,
	}

	for name, ctrl := range controllers {
		if ctrl == nil {
			klog.V(4).Infof("%s is not going to run due to dependent component disabled.", name)
			continue
		}
		if err := mgr.Add(ctrl); err != nil {
			klog.Error(err, "add controller to manager failed", "name", name)
			return err
		}
	}

	// Start cache data after all informer is registered
	klog.V(0).Info("Starting cache resource from apiserver...")
	informerFactory.Start(stopCh)

	// Start image event interface
	klog.V(0).Info("Starting image event collect...")
	imageEventCollect.Start(stopCh)

	// webhook
	hookServer := mgr.GetWebhookServer()
	hookServer.Register("/webhook-v1alpha1-harbor-image", harbor.NewImageEventWebHook(imageEventCollect))
	hookServer.Register("/webhook-v1alpha1-pod-latest-tag", &webhook.Admission{Handler: latesttag.NewLatestTagWebHook(repositoryService)})

	klog.V(0).Info("Starting the controllers.")
	if err = mgr.Start(stopCh); err != nil {
		klog.Fatalf("unable to run the manager: %v", err)
	}
	return nil
}
