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

	"github.com/jmckind/rethinkdb-operator/pkg/apis/operator/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockCluster struct {
	mock.Mock
}

func (m *MockCluster) CreateOrUpdateClusterService() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCluster) CreateOrUpdateDriverService() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCluster) CreateOrUpdateStatefulSet() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCluster) SetDefaults() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCluster) UpdateStatus() error {
	args := m.Called()
	return args.Error(0)
}

type ClusterTestSuite struct {
	suite.Suite
}

func (suite *ClusterTestSuite) SetupTest() {
	// Run before each test...
}

func (suite *ClusterTestSuite) TestCreateClusterWithNilResource() {
	cluster := NewRethinkDBCluster(nil)
	assert.Nil(suite.T(), cluster.Resource)
}

func (suite *ClusterTestSuite) TestCreateClusterWithValidInput() {
	cluster := NewRethinkDBCluster(&v1alpha1.RethinkDB{})
	assert.NotNil(suite.T(), cluster.Resource)
}

// Run test suite...
func TestClusterTestSuite(t *testing.T) {
	suite.Run(t, new(ClusterTestSuite))
}
