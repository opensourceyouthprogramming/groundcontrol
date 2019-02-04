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
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"github.com/stratumn/groundcontrol/relay"
)

// WorkspacesConfig contains all the data in a YAML workspaces config file.
type WorkspacesConfig struct {
	Filename   string `json:"-" yaml:"-"`
	Workspaces []struct {
		Slug        string  `json:"slug"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Notes       *string `json:"notes"`
		Projects    []struct {
			Slug        string  `json:"slug"`
			Repository  string  `json:"repository"`
			Branch      string  `json:"branch"`
			Description *string `json:"description"`
		} `json:"projects" yaml:",flow"`
		Tasks []struct {
			Name  string `json:"name"`
			Steps []struct {
				Projects []string `json:"projects"`
				Commands []string `json:"commands"`
			} `json:"tasks"`
		} `json:"tasks"`
	} `json:"workspaces"`
}

// UpsertNodes upserts nodes for the content of the config.
// It returns the IDs of the workspaces creates.
func (c WorkspacesConfig) UpsertNodes(nodes *NodeManager) ([]string, error) {
	var workspaceIDs []string

	for _, workspaceConfig := range c.Workspaces {
		workspace := Workspace{
			ID:          relay.EncodeID(NodeTypeWorkspace, workspaceConfig.Slug),
			Slug:        workspaceConfig.Slug,
			Name:        workspaceConfig.Name,
			Description: workspaceConfig.Description,
			Notes:       workspaceConfig.Notes,
		}

		projectSlugToID := map[string]string{}

		for _, projectConfig := range workspaceConfig.Projects {
			project := Project{
				ID: relay.EncodeID(
					NodeTypeProject,
					workspace.Slug,
					projectConfig.Slug,
				),
				Slug:        projectConfig.Slug,
				Repository:  projectConfig.Repository,
				Branch:      projectConfig.Branch,
				Description: projectConfig.Description,
				WorkspaceID: workspace.ID,
			}

			nodes.MustStoreProject(project)
			workspace.ProjectIDs = append(workspace.ProjectIDs, project.ID)
			projectSlugToID[project.Slug] = project.ID
		}

		for i, taskConfig := range workspaceConfig.Tasks {
			task := Task{
				ID: relay.EncodeID(
					NodeTypeTask,
					workspace.Slug,
					fmt.Sprint(i),
				),
				Name:        taskConfig.Name,
				WorkspaceID: workspace.ID,
			}

			for j, stepConfig := range taskConfig.Steps {
				var projectIDs []string
				var commandIDs []string

				for _, slug := range stepConfig.Projects {
					id, ok := projectSlugToID[slug]
					if !ok {
						return nil, ErrNotFound
					}
					projectIDs = append(projectIDs, id)
				}

				for k, command := range stepConfig.Commands {
					id := relay.EncodeID(
						NodeTypeCommand,
						workspace.Slug,
						fmt.Sprint(i),
						fmt.Sprint(j),
						fmt.Sprint(k),
					)
					nodes.MustStoreCommand(Command{
						ID:      id,
						Command: command,
					})
					commandIDs = append(commandIDs, id)
				}

				step := Step{
					ID: relay.EncodeID(
						NodeTypeStep,
						workspace.Slug,
						fmt.Sprint(i),
						fmt.Sprint(j),
					),
					ProjectIDs: projectIDs,
					CommandIDs: commandIDs,
					TaskID:     task.ID,
				}

				nodes.MustStoreStep(step)
				task.StepIDs = append(task.StepIDs, step.ID)
			}

			nodes.MustStoreTask(task)
			workspace.TaskIDs = append(workspace.TaskIDs, task.ID)
		}

		nodes.MustStoreWorkspace(workspace)
		workspaceIDs = append(workspaceIDs, workspace.ID)
	}

	return workspaceIDs, nil
}

// LoadWorkspacesConfigYAML loads a config from a YAML file.
func LoadWorkspacesConfigYAML(filename string) (WorkspacesConfig, error) {
	config := WorkspacesConfig{
		Filename: filename,
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = yaml.UnmarshalStrict(bytes, &config)

	return config, err
}