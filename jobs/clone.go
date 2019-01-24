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

package jobs

import (
	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/pubsub"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Clone clones a remote repository locally.
func Clone(
	nodes *models.NodeManager,
	jobs *models.JobManager,
	subs *pubsub.PubSub,
	getProjectPath models.ProjectPathGetter,
	projectID string,
) (string, error) {
	var (
		err         error
		workspaceID string
	)

	err = nodes.LockProject(projectID, func(project models.Project) {
		if project.IsCloning {
			err = ErrDuplicate
			return
		}

		workspaceID = project.WorkspaceID
		project.IsCloning = true
		nodes.MustStoreProject(project)
	})
	if err != nil {
		return "", err
	}

	subs.Publish(models.ProjectUpdated, projectID)
	subs.Publish(models.WorkspaceUpdated, workspaceID)

	jobID := jobs.Add(CloneJob, projectID, func() error {
		return doClone(
			nodes,
			subs,
			getProjectPath,
			projectID,
			workspaceID,
		)
	})

	return jobID, nil
}

func doClone(
	nodes *models.NodeManager,
	subs *pubsub.PubSub,
	getProjectPath models.ProjectPathGetter,
	projectID string,
	workspaceID string,
) error {
	project := nodes.MustLoadProject(projectID)

	if project.IsCloned(nodes, getProjectPath) {
		return ErrCloned
	}

	defer func() {
		nodes.MustLockProject(projectID, func(project models.Project) {
			project.IsCloning = false
			nodes.MustStoreProject(project)
		})

		subs.Publish(models.ProjectUpdated, projectID)
		subs.Publish(models.WorkspaceUpdated, workspaceID)
	}()

	workspace := project.Workspace(nodes)
	directory := getProjectPath(workspace.Slug, project.Repository, project.Branch)

	_, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:           project.Repository,
		ReferenceName: plumbing.NewBranchReferenceName(project.Branch),
	})

	return err
}
