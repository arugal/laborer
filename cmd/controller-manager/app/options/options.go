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

package options

import (
	"flag"
	"strings"
	"time"

	repositoryservice "github.com/arugal/laborer/pkg/service/repository"
	"github.com/arugal/laborer/pkg/simple/client/k8s"
	"github.com/spf13/pflag"
	"k8s.io/klog"

	"k8s.io/client-go/tools/leaderelection"
	cliflag "k8s.io/component-base/cli/flag"
)

type LaborerControllerManagerOptions struct {
	KubernetesOptions        *k8s.KubernetesOptions
	LeaderElect              bool
	LeaderElectNamespace     string
	LeaderElection           *leaderelection.LeaderElectionConfig
	WebhookCertDir           string
	RepositoryServiceOptions *repositoryservice.RepositoryServiceOptions
}

func NewLaborerControllerManagerOptions() *LaborerControllerManagerOptions {
	return &LaborerControllerManagerOptions{
		KubernetesOptions: k8s.NewKubernetesOptions(),
		LeaderElection: &leaderelection.LeaderElectionConfig{
			LeaseDuration: 30 * time.Second,
			RenewDeadline: 15 * time.Second,
			RetryPeriod:   5 * time.Second,
		},
		LeaderElect:              false,
		LeaderElectNamespace:     "",
		WebhookCertDir:           "",
		RepositoryServiceOptions: repositoryservice.NewRepositoryServiceOptions(),
	}
}

func (s *LaborerControllerManagerOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}

	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)
	s.RepositoryServiceOptions.AddFlags(fss.FlagSet("repository"))

	fs := fss.FlagSet("leaderelection")
	s.bindLeaderElectionFlags(s.LeaderElection, fs)

	fs.BoolVar(&s.LeaderElect, "leader-elect", s.LeaderElect, ""+
		"Whether to enable leader election. This field should be enabled when controller manager"+
		"deployed with multiple replicas.")

	fs.StringVar(&s.LeaderElectNamespace, "leader-elect-namespace", s.LeaderElectNamespace, ""+
		"Determines the namespace in which the leader election configmap will be created.")

	fs.StringVar(&s.WebhookCertDir, "webhook-cert-dir", s.WebhookCertDir, ""+
		"Certificate directory used to setup webhooks, need tls.crt and tls.key placed inside."+
		"if not set, webhook server would look up the server key and certificate in"+
		"{TempDir}/k8s-webhook-server/serving-certs")

	kfs := fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		kfs.AddGoFlag(fl)
	})

	return fss
}

func (s *LaborerControllerManagerOptions) Validate() (errs []error) {
	errs = append(errs, s.KubernetesOptions.Validate()...)
	errs = append(errs, s.RepositoryServiceOptions.Validate()...)
	return errs
}

func (s *LaborerControllerManagerOptions) bindLeaderElectionFlags(l *leaderelection.LeaderElectionConfig, fs *pflag.FlagSet) {
	fs.DurationVar(&l.LeaseDuration, "leader-elect-lease-duration", l.LeaseDuration, ""+
		"The duration that non-leader candidates will wait after observing a leadership "+
		"renewal until attempting to acquire leadership of a led but unrenewed leader "+
		"slot. This is effectively the maximum duration that a leader can be stopped "+
		"before it is replaced by another candidate. This is only applicable if leader "+
		"election is enabled.")
	fs.DurationVar(&l.RenewDeadline, "leader-elect-renew-deadline", l.RenewDeadline, ""+
		"The interval between attempts by the acting master to renew a leadership slot "+
		"before it stops leading. This must be less than or equal to the lease duration. "+
		"This is only applicable if leader election is enabled.")
	fs.DurationVar(&l.RetryPeriod, "leader-elect-retry-period", l.RetryPeriod, ""+
		"The duration the clients should wait between attempting acquisition and renewal "+
		"of a leadership. This is only applicable if leader election is enabled.")
}
