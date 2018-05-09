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

package main

import (
	"context"
	"net/http"
	"runtime"

	operator "github.com/jmckind/rethinkdb-operator/pkg/stub"
	operatorVersion "github.com/jmckind/rethinkdb-operator/version"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	"github.com/jmckind/rethinkdb-operator/pkg/util/probe"
	"github.com/sirupsen/logrus"
)

func main() {
	printVersion()
	startReadyz()
	run()
}

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
	logrus.Infof("rethinkdb-operator Version: %v", operatorVersion.Version)
}

func startReadyz() {
	logrus.Info("Starting readyz endpoint...")
	http.HandleFunc(probe.HTTPReadyzEndpoint, probe.ReadyzHandler)
	go http.ListenAndServe("0.0.0.0:8080", nil)
}

func run() {
	sdk.Watch("operator.rethinkdb.com/v1alpha1", "RethinkDB", "default", 5)
	sdk.Handle(operator.NewRethinkDBHandler())
	sdk.Run(context.TODO())
}
