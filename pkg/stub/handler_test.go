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
	"errors"
	"testing"

	v1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/operator/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
}

func (suite *HandlerTestSuite) SetupTest() {
	// Run before each test...
}

func (suite *HandlerTestSuite) TestHandleWithNilObject() {
	context := types.Context{}
	event := types.Event{}
	assert.Nil(suite.T(), event.Object)

	handler := NewRethinkDBHandler()
	err := handler.Handle(context, event)
	assert.Nil(suite.T(), err)
}

func (suite *HandlerTestSuite) TestHandleWithDefaultRethinkDB() {
	context := types.Context{}
	event := types.Event{Object: &v1alpha1.RethinkDB{}}

	handler := NewRethinkDBHandler()
	err := handler.Handle(context, event)

	assert.Error(suite.T(), err)
}

func (suite *HandlerTestSuite) TestHandleRethinkDBWithNilInput() {
	handler := NewRethinkDBHandler()
	err := handler.HandleRethinkDB(nil)
	assert.Error(suite.T(), err)
}

func (suite *HandlerTestSuite) TestHandleRethinkDBWithValidInput() {
	cluster := new(MockCluster)
	cluster.On("SetDefaults").Return(false)
	cluster.On("CreateOrUpdateClusterService").Return(nil)
	cluster.On("CreateOrUpdateDriverService").Return(nil)
	cluster.On("CreateOrUpdateStatefulSet").Return(nil)
	cluster.On("UpdateStatus").Return(nil)

	handler := NewRethinkDBHandler()
	err := handler.HandleRethinkDB(cluster)

	cluster.AssertExpectations(suite.T())
	assert.Nil(suite.T(), err)
}

func (suite *HandlerTestSuite) TestHandleRethinkDBWithClusterServiceFailure() {
	cluster := new(MockCluster)
	cluster.On("SetDefaults").Return(false)
	cluster.On("CreateOrUpdateClusterService").Return(errors.New("failed"))

	handler := NewRethinkDBHandler()
	err := handler.HandleRethinkDB(cluster)

	cluster.AssertExpectations(suite.T())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "failed to create or update cluster service: failed", err.Error())
}

func (suite *HandlerTestSuite) TestHandleRethinkDBWithDriverServiceFailure() {
	cluster := new(MockCluster)
	cluster.On("SetDefaults").Return(false)
	cluster.On("CreateOrUpdateClusterService").Return(nil)
	cluster.On("CreateOrUpdateDriverService").Return(errors.New("failed"))

	handler := NewRethinkDBHandler()
	err := handler.HandleRethinkDB(cluster)

	cluster.AssertExpectations(suite.T())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "failed to create or update driver service: failed", err.Error())
}

func (suite *HandlerTestSuite) TestHandleRethinkDBWithStatefulSetFailure() {
	cluster := new(MockCluster)
	cluster.On("SetDefaults").Return(false)
	cluster.On("CreateOrUpdateClusterService").Return(nil)
	cluster.On("CreateOrUpdateDriverService").Return(nil)
	cluster.On("CreateOrUpdateStatefulSet").Return(errors.New("failed"))

	handler := NewRethinkDBHandler()
	err := handler.HandleRethinkDB(cluster)

	cluster.AssertExpectations(suite.T())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "failed to create or update statefulset: failed", err.Error())
}

func (suite *HandlerTestSuite) TestHandleRethinkDBWithUpdateStatusFailure() {
	cluster := new(MockCluster)
	cluster.On("SetDefaults").Return(false)
	cluster.On("CreateOrUpdateClusterService").Return(nil)
	cluster.On("CreateOrUpdateDriverService").Return(nil)
	cluster.On("CreateOrUpdateStatefulSet").Return(nil)
	cluster.On("UpdateStatus").Return(errors.New("failed"))

	handler := NewRethinkDBHandler()
	err := handler.HandleRethinkDB(cluster)

	cluster.AssertExpectations(suite.T())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "failed to update rethinkdb status: failed", err.Error())
}

// Run test suite...
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
