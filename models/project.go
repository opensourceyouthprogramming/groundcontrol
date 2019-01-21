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

package models

import (
	"container/list"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/stratumn/groundcontrol/date"

	"github.com/stratumn/groundcontrol/relay"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type Project struct {
	ID          string     `json:"id"`
	Repository  string     `json:"repository"`
	Branch      string     `json:"branch"`
	Description *string    `json:"description"`
	IsCloning   bool       `json:"isCloning"`
	IsCloned    bool       `json:"isCloned"`
	Workspace   *Workspace `json:"workspace"`

	commitList *list.List

	commitsMu        sync.Mutex
	isLoadingCommits bool

	cloneMu   sync.Mutex
	isCloning bool
}

func (*Project) IsNode() {}

var commitPaginator = relay.Paginator{
	GetID: func(node interface{}) string {
		return node.(Commit).ID
	},
}

func (p *Project) Commits(after, before *string, first, last *int) (*CommitConnection, error) {
	if p.commitList.Len() == 0 {
		p.commitsMu.Lock()
		defer p.commitsMu.Unlock()

		if !p.isLoadingCommits {
			p.isLoadingCommits = true

			CreateJob(
				"Load Commits",
				p,
				func() error {
					err := p.loadCommits()
					p.commitsMu.Lock()
					p.isLoadingCommits = false
					p.commitsMu.Unlock()
					PublishProjectUpdated(p)
					PublishWorkspaceUpdated(p.Workspace)
					return err
				},
			)
		}

		return &CommitConnection{
			IsLoading: true,
		}, nil
	}

	connection, err := commitPaginator.Paginate(p.commitList, after, before, first, last)
	if err != nil {
		return nil, err
	}

	edges := make([]CommitEdge, len(connection.Edges))

	for i, v := range connection.Edges {
		edges[i] = CommitEdge{
			Node:   v.Node.(Commit),
			Cursor: v.Cursor,
		}
	}

	return &CommitConnection{
		Edges:     edges,
		PageInfo:  connection.PageInfo,
		IsLoading: p.isLoadingCommits,
	}, nil
}

func (p *Project) loadCommits() error {
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           p.Repository,
		ReferenceName: plumbing.NewBranchReferenceName(p.Branch),
	})
	if err != nil {
		return err
	}

	ref, err := repo.Head()
	if err != nil {
		return err
	}

	iter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}

	return iter.ForEach(func(c *object.Commit) error {
		p.commitList.PushBack(Commit{
			ID:       c.Hash.String(),
			Headline: strings.Split(c.Message, "\n")[0],
			Message:  c.Message,
			Author:   c.Author.Name,
			Date:     c.Author.When.Format(date.DateFormat),
		})

		return nil
	})
}

func (p *Project) findCommitElement(id string) *list.Element {
	element := p.commitList.Front()

	for element != nil {
		if element.Value.(Commit).ID == id {
			return element
		}
		element = element.Next()
	}

	return nil
}

var (
	nextProjectSubscriptionID   = uint64(0)
	projectUpdatedSubscriptions = sync.Map{}
)

func SubscribeProjectUpdated(fn func(*Project)) func() {
	id := atomic.AddUint64(&nextProjectSubscriptionID, 1)
	projectUpdatedSubscriptions.Store(id, fn)

	return func() {
		projectUpdatedSubscriptions.Delete(id)
	}
}

func PublishProjectUpdated(project *Project) {
	projectUpdatedSubscriptions.Range(func(_, v interface{}) bool {
		fn := v.(func(*Project))
		fn(project)
		return true
	})
}