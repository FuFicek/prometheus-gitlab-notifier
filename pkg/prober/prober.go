// Copyright 2019 FUSAKLA Martin Chodúr
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prober

import (
	"io"
	"net/http"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
)

// NewInRouter returns new Prober which registers it's endpoints in the Router to provide readiness and liveness endpoints.
func NewInRouter(logger log.Logger, router *mux.Router) *prober {
	p := &prober{
		logger:      logger,
		serverReady: nil,
	}
	p.registerInRouter(router)
	return p
}

// prober holds application readiness/liveness status and provides handlers for reporting it.
type prober struct {
	logger         log.Logger
	serverReadyMtx sync.RWMutex
	serverReady    error
}

func (p *prober) registerInRouter(router *mux.Router) {
	router.HandleFunc("/liveness", p.livenessHandler)
	router.HandleFunc("/readiness", p.readinessHandler)
}

func (p *prober) livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `OK`)
}

func (p *prober) writeFailedReadiness(w http.ResponseWriter, err error) {
	level.Error(p.logger).Log("msg", "readiness probe failed", "err", err)
	http.Error(w, err.Error(), http.StatusServiceUnavailable)
}

func (p *prober) readinessHandler(w http.ResponseWriter, r *http.Request) {
	if err := p.isReady(); err != nil {
		p.writeFailedReadiness(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `OK`)
}

// SetServerNotReady sets the readiness probe to invalid state.
func (p *prober) SetServerNotReady(err error) {
	p.serverReadyMtx.Lock()
	defer p.serverReadyMtx.Unlock()
	level.Warn(p.logger).Log("msg", "Marking server as not ready", "reason", err)
	p.serverReady = err
}

func (p *prober) isReady() error {
	p.serverReadyMtx.RLock()
	defer p.serverReadyMtx.RUnlock()
	return p.serverReady
}
