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

package groundcontrol

// Resolver is the root GraphQL resolver.
type Resolver struct {
	Viewer User
}

// Query returns the resolver for queries.
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// Mutation returns the resolver for mutations.
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Subscription returns the resolver for subscriptions.
func (r *Resolver) Subscription() SubscriptionResolver {
	return &subscriptionResolver{r}
}
