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

package resolvers

import (
	"context"

	"groundcontrol/models"
)

func (r *mutationResolver) SetKey(
	ctx context.Context,
	input models.KeyInput,
) (models.Key, error) {
	modelCtx := models.GetModelContext(ctx)

	id := modelCtx.Keys.UpsertKey(
		modelCtx.Nodes,
		modelCtx.Subs,
		modelCtx.ViewerID,
		input,
	)

	if err := modelCtx.Keys.Save(); err != nil {
		return models.Key{}, err
	}

	return modelCtx.Nodes.MustLoadKey(id), nil
}
