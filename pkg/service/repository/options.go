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

package repository

import (
	"fmt"

	"github.com/spf13/pflag"
)

type RepositoryServiceOptions struct {
	Mock     bool              `json:"mock,omitempty" yaml:"mock,omitempty"`
	MockTags map[string]string `json:"mockTags,omitempty" yaml:"mockTags,omitempty"`
	// repository address
	Host               string `json:"host" yaml:"host"`
	Protocol           string `json:"protocol" yaml:"protocol"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`
	ApiPathPrefix      string `json:"apiPathPrefix" yaml:"apiPathPrefix"`
}

func (r *RepositoryServiceOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&r.Mock, "repository-mock", r.Mock, "use mock repository service")
	fs.StringVar(&r.Host, "repository-host", r.Host, "image repository host, eg: demo.goharbor.io (harbor is currently supported only)")
	fs.StringVar(&r.Protocol, "repository-protocol", r.Protocol, "repository protocol, optional: http; https")
	fs.BoolVar(&r.InsecureSkipVerify, "repository-insecure-skip-verify", r.InsecureSkipVerify,
		"if true, server-side certificate authentication is skipped when the protocol is https")
	fs.StringVar(&r.ApiPathPrefix, "repository-api-path-prefix", r.ApiPathPrefix, "")
}

func (r *RepositoryServiceOptions) Validate() (errs []error) {
	if r.Protocol != "http" && r.Protocol != "https" {
		errs = append(errs, fmt.Errorf("repository protocol only support http, https"))
	}
	return errs
}

func NewRepositoryServiceOptions() *RepositoryServiceOptions {
	return &RepositoryServiceOptions{
		Mock:               false,
		Host:               "",
		Protocol:           "https",
		InsecureSkipVerify: true,
		ApiPathPrefix:      "/api/v2.0",
	}
}
