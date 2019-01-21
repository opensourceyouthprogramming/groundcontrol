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
	"github.com/stratumn/groundcontrol/gql"
	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/pubsub"
)

// Resolver is the root GraphQL resolver.
type Resolver struct {
	Viewer     models.User
	JobManager *models.JobManager
	PubSub     *pubsub.PubSub
}

// Query returns the resolver for queries.
func (r *Resolver) Query() gql.QueryResolver {
	return &queryResolver{r}
}

// Mutation returns the resolver for mutations.
func (r *Resolver) Mutation() gql.MutationResolver {
	return &mutationResolver{r}
}

// Subscription returns the resolver for subscriptions.
func (r *Resolver) Subscription() gql.SubscriptionResolver {
	return &subscriptionResolver{r}
}

// User returns the resolver for a user.
func (r *Resolver) User() gql.UserResolver {
	return &userResolver{r}
}

// Project returns the resolver for a project.
func (r *Resolver) Project() gql.ProjectResolver {
	return &projectResolver{r}
}
