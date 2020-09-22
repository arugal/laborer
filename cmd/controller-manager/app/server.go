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
	"github.com/arugal/laborer/pkg/utils/term"
	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
)

func NewControllerManagerCommand() *cobra.Command {
	s := options.NewLaborerControllerManagerOptions()
	conf, err := config.TryLoadFromDisk()
	if err != nil {
		klog.Fatal("Failed to load configuration from disk", err)
	} else {
		s = &options.LaborerControllerManagerOptions{
			KubernetesOptions: conf.KubernetesOptions,
			LeaderElection:    s.LeaderElection,
			LeaderElect:       s.LeaderElect,
			WebhookCertDir:    s.WebhookCertDir,
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

			// TODO run
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
