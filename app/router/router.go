/*
 * Copyright 2022 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package router

import (
	"fmt"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/rs/cors"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"net/http"
)

const (
	apiV1 = "/api/v1"
)

var (
	c = cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})

	middlewares = chi.Middlewares{
		render.SetContentType(render.ContentTypeJSON),
		middleware.AllowContentType("application/json"),
		middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: config.GetLoggerFor("Validator API Handler"), NoColor: true}),
		middleware.RedirectSlashes,
		middleware.Recoverer,
		middleware.NoCache,
		c.Handler,
	}
)

type APIRouter struct {
	Router *chi.Mux
}

func NewAPIRouter() *APIRouter {
	router := chi.NewRouter()

	router.Use(middlewares...)

	return &APIRouter{
		Router: router,
	}
}

func (api *APIRouter) AddV1Router(path string, router http.Handler) {
	api.Router.Mount(fmt.Sprint(apiV1, path), router)
}
