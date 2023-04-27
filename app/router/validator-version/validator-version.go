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

package validator_version

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"net/http"
	"os"
)

type VersionResponse struct {
	Version string `json:"version"`
}

var (
	Route = "/version"
)

// Router for version check
func NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", versionResponse())
	return r
}

// GET: .../version
func versionResponse() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		version := os.Getenv("VTAG")
		render.JSON(w, r, &VersionResponse{Version: version})
	}
}
