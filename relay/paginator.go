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

package relay

import (
	"container/list"
	"errors"
)

// Pagination errors.
var (
	ErrPaginateFirst = errors.New("first cannot be negative")
	ErrPaginateLast  = errors.New("last cannot be negative")
)

// Paginator helps paginate lists for Relay.
//
// See: https://facebook.github.io/relay/graphql/connections.htm
type Paginator struct {
	// GetID must return the ID of a list value.
	GetID func(node interface{}) string
}

// PaginationConnection represents the result of a pagination.
type PaginationConnection struct {
	Edges    []PaginationEdge
	PageInfo PageInfo
}

// PaginationEdge represents an edge in a pagination connection.
type PaginationEdge struct {
	Cursor string
	Node   interface{}
}

// PageInfo contains fields related to pagination.
type PageInfo struct {
	HasNextPage     *bool   `json:"hasNextPage"`
	HasPreviousPage *bool   `json:"hasPreviousPage"`
	EndCursor       *string `json:"endCursor"`
	StartCursor     *string `json:"startCursor"`
}

// Paginate paginates a list given query parameters.
func (p Paginator) Paginate(l *list.List, after, before *string, first, last *int) (*PaginationConnection, error) {
	edges, hadMore := p.applyCursors(l, after, before)
	pageInfo := PageInfo{
		HasNextPage:     new(bool),
		HasPreviousPage: new(bool),
		EndCursor:       new(string),
		StartCursor:     new(string),
	}

	*pageInfo.HasNextPage = false
	*pageInfo.HasPreviousPage = false

	if first != nil {
		firstValue := *first
		if firstValue < 0 {
			return nil, ErrPaginateFirst
		}
		if firstValue > len(edges) {
			firstValue = len(edges)
		}
		if firstValue < len(edges) {
			*pageInfo.HasNextPage = true
		} else if before != nil {
			*pageInfo.HasNextPage = hadMore
		}
		edges = edges[:firstValue]
	} else if before != nil {
		*pageInfo.HasNextPage = hadMore
	}

	if last != nil {
		lastValue := *last
		if lastValue < 0 {
			return nil, ErrPaginateLast
		}
		if lastValue > len(edges) {
			lastValue = len(edges)
		}
		if lastValue < len(edges) {
			*pageInfo.HasPreviousPage = true
		} else if after != nil {
			*pageInfo.HasPreviousPage = hadMore
		}
		end := len(edges) - lastValue
		edges = edges[end:]
	} else if after != nil {
		*pageInfo.HasPreviousPage = hadMore
	}

	if len(edges) > 0 {
		*pageInfo.EndCursor = edges[len(edges)-1].Cursor
		*pageInfo.StartCursor = edges[0].Cursor
	}

	return &PaginationConnection{
		Edges:    edges,
		PageInfo: pageInfo,
	}, nil
}

func (p Paginator) applyCursors(l *list.List, after, before *string) (edges []PaginationEdge, hadMore bool) {
	if after != nil {
		element := p.find(l, *after)
		if element != nil {
			hadMore = element.Prev() != nil
			element = element.Next()
			for element != nil {
				edges = append(edges, PaginationEdge{
					Cursor: p.GetID(element.Value),
					Node:   element.Value,
				})
				element = element.Next()
			}
		}
		return
	}

	if before != nil {
		element := p.find(l, *before)
		if element != nil {
			hadMore = element.Next() != nil
			element = l.Front()
			for element != nil {
				id := p.GetID(element.Value)
				if before != nil && *before == id {
					return nil, hadMore
				}
				edges = append(edges, PaginationEdge{
					Cursor: p.GetID(element.Value),
					Node:   element.Value,
				})
				element = element.Next()
			}
		}
		return
	}

	element := l.Front()
	for element != nil {
		edges = append(edges, PaginationEdge{
			Cursor: p.GetID(element.Value),
			Node:   element.Value,
		})
		element = element.Next()
	}

	return
}

func (p Paginator) find(l *list.List, id string) *list.Element {
	element := l.Front()

	for element != nil {
		if p.GetID(element.Value) == id {
			return element
		}
		element = element.Next()
	}

	return nil
}