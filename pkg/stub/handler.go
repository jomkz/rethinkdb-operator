// Copyright 2018 The rethinkdb-operator Authors
//
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

package stub

import (
	"github.com/coreos/operator-sdk/pkg/sdk/handler"
	"github.com/coreos/operator-sdk/pkg/sdk/types"
	v1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/operator/v1alpha1"
	"github.com/sirupsen/logrus"
)

const (
	defaultConfig = `bind=all
directory=/var/lib/rethinkdb/default
`
)

type RethinkDBHandler struct {
	// TODO...
}

func NewRethinkDBHandler() handler.Handler {
	return &RethinkDBHandler{}
}

func (h *RethinkDBHandler) Handle(ctx types.Context, event types.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.RethinkDB:
		return h.HandleRethinkDB(o)
	}
	return nil
}

func (h *RethinkDBHandler) HandleRethinkDB(r *v1alpha1.RethinkDB) error {
	logrus.Infof("handling rethinkdb: %v", r.Name)
	r.SetDefaults()
	cluster := NewRethinkDBCluster(r)
	return cluster.CreateOrUpdateCluster()
}
