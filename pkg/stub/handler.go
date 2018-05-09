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
	"fmt"

	v1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/operator/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk/handler"
	"github.com/operator-framework/operator-sdk/pkg/sdk/types"
)

const (
	defaultConfig = `bind=all
directory=/var/lib/rethinkdb/default
`
)

type RethinkDBHandler struct {
	handler.Handler
}

func NewRethinkDBHandler() *RethinkDBHandler {
	return &RethinkDBHandler{}
}

func (h *RethinkDBHandler) Handle(ctx types.Context, event types.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.RethinkDB:
		return h.HandleRethinkDB(NewRethinkDBCluster(o))
	}
	return nil
}

func (h *RethinkDBHandler) HandleRethinkDB(c RethinkDBController) error {
	if c == nil {
		return fmt.Errorf("controller cannot be nil")
	}

	c.SetDefaults()

	err := c.CreateOrUpdateClusterService()
	if err != nil {
		return fmt.Errorf("failed to create or update cluster service: %v", err)
	}

	err = c.CreateOrUpdateDriverService()
	if err != nil {
		return fmt.Errorf("failed to create or update driver service: %v", err)
	}

	err = c.CreateOrUpdateStatefulSet()
	if err != nil {
		return fmt.Errorf("failed to create or update statefulset: %v", err)
	}

	err = c.UpdateStatus()
	if err != nil {
		return fmt.Errorf("failed to update rethinkdb status: %v", err)
	}

	return nil
}
