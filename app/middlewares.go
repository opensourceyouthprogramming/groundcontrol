// Copyright 2019 Stratumn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"

	"groundcontrol/models"
)

func modelContextResolverMiddleware(modelCtx *models.ModelContext) graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
		return next(models.WithModelContext(ctx, modelCtx))
	}
}

func logHTTPRequestMiddleware(log *models.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug("HTTP request %s %s %s", r.Method, r.URL.String(), r.RemoteAddr)
			h.ServeHTTP(w, r)
		})
	}
}
