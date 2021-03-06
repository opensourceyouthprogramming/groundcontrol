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
	"context"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"groundcontrol/models"
)

// Clone clones a remote repository locally.
func Clone(ctx context.Context, projectID string, priority models.JobPriority) (string, error) {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	subs := modelCtx.Subs
	workspaceID := ""

	err := nodes.LockProjectE(projectID, func(project models.Project) error {
		if project.IsCloning {
			return ErrDuplicate
		}

		workspaceID = project.WorkspaceID
		project.IsCloning = true
		nodes.MustStoreProject(project)

		return nil
	})
	if err != nil {
		return "", err
	}

	subs.Publish(models.ProjectUpserted, projectID)
	subs.Publish(models.WorkspaceUpserted, workspaceID)

	jobID := modelCtx.Jobs.Add(
		models.GetModelContext(ctx),
		CloneJob,
		projectID,
		priority,
		func(ctx context.Context) error {
			return doClone(ctx, projectID, workspaceID)
		},
	)

	return jobID, nil
}

func doClone(ctx context.Context, projectID string, workspaceID string) error {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	subs := modelCtx.Subs
	project := nodes.MustLoadProject(projectID)

	if project.IsCloned(ctx) {
		return ErrCloned
	}

	defer func() {
		nodes.MustLockProject(projectID, func(project models.Project) {
			project.IsCloning = false
			nodes.MustStoreProject(project)
		})

		subs.Publish(models.ProjectUpserted, projectID)
		subs.Publish(models.WorkspaceUpserted, workspaceID)
	}()

	workspace := project.Workspace(ctx)
	directory := modelCtx.GetProjectPath(workspace.Slug, project.Repository, project.Branch)

	_, err := git.PlainCloneContext(
		ctx,
		directory,
		false,
		&git.CloneOptions{
			URL:           project.Repository,
			ReferenceName: plumbing.NewBranchReferenceName(project.Branch),
		},
	)

	return err
}
