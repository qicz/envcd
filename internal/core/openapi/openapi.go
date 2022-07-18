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
	"net/http"
	"time"

	"github.com/acmestack/envcd/internal/core/plugin"
	"github.com/acmestack/envcd/internal/core/plugin/logging"
	"github.com/acmestack/envcd/internal/core/plugin/permission"
	"github.com/acmestack/envcd/internal/core/plugin/response"
	"github.com/acmestack/envcd/internal/core/service/routers"
	"github.com/acmestack/envcd/internal/core/storage"
	"github.com/acmestack/envcd/internal/envcd"
	"github.com/acmestack/envcd/internal/pkg/config"
	"github.com/acmestack/envcd/internal/pkg/context"
	"github.com/acmestack/envcd/internal/pkg/executor"
	"github.com/acmestack/envcd/pkg/entity/data"
	"github.com/acmestack/godkits/gox/errorsx"
	"github.com/acmestack/godkits/log"
	"github.com/gin-gonic/gin"
)

type Openapi struct {
	envcd     *envcd.Envcd
	storage   *storage.Storage
	executors []executor.Executor
}

func Start(serverSetting *config.Server, envcd *envcd.Envcd, storage *storage.Storage) {
	openapi := &Openapi{
		envcd:     envcd,
		storage:   storage,
		executors: []executor.Executor{logging.New(), permission.New(), response.New()},
	}
	// sort plugin
	plugin.Sort(openapi.executors)
	openapi.initServer(serverSetting)
	openapi.openRouter()
}

func (openapi *Openapi) initServer(serverSetting *config.Server) {
	gin.SetMode(serverSetting.RunMode)

	routersInit := routers.InitRouter()
	readTimeout := serverSetting.ReadTimeout
	writeTimeout := serverSetting.WriteTimeout
	endPoint := fmt.Sprintf(":%d", serverSetting.HttpPort)
	maxHeaderBytes := 1 << 20

	server := &http.Server{
		Addr:           endPoint,
		Handler:        routersInit,
		ReadTimeout:    time.Duration(readTimeout),
		WriteTimeout:   time.Duration(writeTimeout),
		MaxHeaderBytes: maxHeaderBytes,
	}

	log.Info("[info] start http server listening %s", endPoint)

	err := server.ListenAndServe()
	if err != nil {
		log.Error("service error %v", err)
		return
	}
}

// todo open Router
func (openapi *Openapi) openRouter() {
	// fixme: plugin.NewChain(openapi.executors) for peer request
	// plugin.NewChain(openapi.executors)
	c := &context.Context{Action: func() (*data.EnvcdResult, error) {
		fmt.Println("hello world")
		// openapi.envcd.Put("key", "value")
		return nil, errorsx.Err("test error")
	}}
	if ret, err := plugin.NewChain(openapi.executors).Execute(c); err != nil {
		fmt.Printf("ret = %v, error = %v", ret, err)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// user auth
	r.POST("/login", openapi.logins())
	r.Run()
}
