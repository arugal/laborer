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

package verify

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/runtime"
)

var (
	expect  = ""
	request = ""
	retries = 1
)

func httpVerifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "Verify that the HTTP response is as expect",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			for i := 0; i <= retries; i++ {
				err = visit(request, expect)
				if err == nil {
					return nil
				}
				// sleep
				time.Sleep(time.Second)
			}

			return err
		},
	}

	cmd.PersistentFlags().StringVar(&expect, "expect", expect, "expect response")
	cmd.PersistentFlags().StringVar(&request, "request", request, "the request url")
	cmd.PersistentFlags().IntVar(&retries, "retries", retries, "number of access retries")

	runtime.Must(cmd.MarkPersistentFlagRequired("expect"))
	runtime.Must(cmd.MarkPersistentFlagRequired("request"))
	return cmd
}

func visit(req, expected string) error {
	client := http.Client{}
	resp, err := client.Get(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	actual := string(body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("the request status code [%d] was unsuccessful, %s", resp.StatusCode, actual)
	}

	if expected == "not null" && len(actual) > 0 {
		return nil
	}

	if expected != actual {
		return fmt.Errorf("actual: %s, expecte： %s", actual, expected)
	}
	return nil
}
