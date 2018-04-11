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
	"testing"

	"github.com/coreos/operator-sdk/pkg/sdk/types"
	v1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/operator/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestHandleWithNilObject(t *testing.T) {
	context := types.Context{}
	event := types.Event{}
	assert.Nil(t, event.Object)

	handler := NewRethinkDBHandler()
	err := handler.Handle(context, event)
	assert.Nil(t, err)
}

func TestHandleWithDefaultRethinkDB(t *testing.T) {
	context := types.Context{}
	event := types.Event{Object: &v1alpha1.RethinkDB{}}

	handler := NewRethinkDBHandler()
	err := handler.Handle(context, event)

	// TODO: Fix this once mocking setup
	assert.NotNil(t, err)
}
