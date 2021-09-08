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

package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"

	"k8s.io/klog"
)

const (
	DefaultPort = 9080
)

type HttpServer struct {
	Host string

	Port int

	serveMux *http.ServeMux

	handlers map[string]http.Handler

	// defaultingOnce ensures that the default fields are only ever set once.
	defaultingOnce sync.Once

	// mu protects access to the handler map for Start, Register, etc
	mu sync.Mutex
}

// NewHttpServer return http server
func NewHttpServer() *HttpServer {
	return &HttpServer{}
}

func (h *HttpServer) setDefaults() {
	h.handlers = map[string]http.Handler{}
	if h.serveMux == nil {
		h.serveMux = http.NewServeMux()
	}

	if h.Port <= 0 {
		h.Port = DefaultPort
	}
}

func (h *HttpServer) Register(path string, handler http.Handler) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.defaultingOnce.Do(h.setDefaults)

	_, found := h.handlers[path]
	if found {
		panic(fmt.Errorf("can't register duplicate path: %v", path))
	}

	h.handlers[path] = handler
	h.serveMux.Handle(path, handler)
	klog.V(0).Infof("%s registering handler", path)
}

func (h *HttpServer) Start(ctx context.Context) error {
	h.defaultingOnce.Do(h.setDefaults)

	klog.V(0).Info("starting http server")

	listener, err := net.Listen("tcp", net.JoinHostPort(h.Host, strconv.Itoa(h.Port)))
	if err != nil {
		return err
	}

	srv := &http.Server{
		Handler: h.serveMux,
	}

	idleConnsClosed := make(chan struct{})

	go func() {
		<-ctx.Done()
		klog.V(0).Info("shutting down http server")

		if err := srv.Shutdown(context.Background()); err != nil {
			klog.Errorf("error shutting down the http server %v", err)
		}
		close(idleConnsClosed)
	}()

	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		return err
	}

	<-idleConnsClosed
	return nil
}
