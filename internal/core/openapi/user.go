/*
 * Copyright (c) 2022, AcmeStack
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package openapi

import (
	"fmt"

	"github.com/acmestack/envcd/internal/core/plugin"
	"github.com/acmestack/envcd/internal/envcd"
	"github.com/acmestack/envcd/internal/pkg/context"
	"github.com/acmestack/envcd/internal/pkg/executor"
	"github.com/acmestack/envcd/pkg/entity/data"
	"github.com/acmestack/godkits/gox/errorsx"
	"github.com/gin-gonic/gin"
)

type UserApi struct {
	Openapi
}

func newUserApi(envcd *envcd.Envcd, executors []executor.Executor) *UserApi {
	api := &UserApi{}
	api.envcd = envcd
	api.executors = executors
	return api
}

func (userapi *UserApi) login(ctx *gin.Context) {
	c := &context.Context{Action: func() (*data.EnvcdResult, error) {
		fmt.Println("hello world")
		userapi.envcd.Put("key", "value")
		return nil, errorsx.Err("test error")
	}}
	if ret, err := plugin.NewChain(userapi.executors).Execute(c); err != nil {
		fmt.Printf("ret = %v, error = %v", ret, err)
	}
}
