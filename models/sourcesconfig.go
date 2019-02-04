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
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/stratumn/groundcontrol/relay"
)

// SourcesConfig contains all the data in a YAML sources config file.
type SourcesConfig struct {
	Filename         string `json:"-" yaml:"-"`
	DirectorySources []struct {
		Directory string `json:"directory"`
	} `json:"directorySources" yaml:"directory-sources"`
	GitSources []struct {
		Repository string `json:"repository"`
		Branch     string `json:"branch"`
	} `json:"gitSources" yaml:"git-sources"`
}

// UpsertNodes upserts nodes for the content of the sources config.
// The user node must already exists
func (c SourcesConfig) UpsertNodes(nodes *NodeManager, userID string) error {
	var sourceIDs []string

	for _, sourceConfig := range c.DirectorySources {
		source := DirectorySource{
			ID:        relay.EncodeID(NodeTypeDirectorySource, sourceConfig.Directory),
			Directory: sourceConfig.Directory,
		}

		nodes.MustStoreDirectorySource(source)
		sourceIDs = append(sourceIDs, source.ID)
	}

	for _, sourceConfig := range c.GitSources {
		source := GitSource{
			ID: relay.EncodeID(
				NodeTypeDirectorySource,
				sourceConfig.Repository,
				sourceConfig.Branch,
			),
			Repository: sourceConfig.Repository,
			Branch:     sourceConfig.Branch,
		}

		nodes.MustStoreGitSource(source)
		sourceIDs = append(sourceIDs, source.ID)
	}

	nodes.MustLockUser(userID, func(user User) {
		user.SourceIDs = sourceIDs
		nodes.MustStoreUser(user)
	})

	return nil
}

// Save saves the config to disk, overwriting the file if it exists.
func (c SourcesConfig) Save() error {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.Filename, bytes, 0644)
}

// LoadSourcesConfigYAML loads a source config from a YAML file.
// It will create a file if it doesn't exist.
func LoadSourcesConfigYAML(filename string) (SourcesConfig, error) {
	config := SourcesConfig{
		Filename: filename,
	}

	bytes, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		config := SourcesConfig{
			Filename: filename,
		}
		if err := config.Save(); err != nil {
			return SourcesConfig{}, err
		}

		return LoadSourcesConfigYAML(filename)
	}
	if err != nil {
		return config, err
	}

	err = yaml.UnmarshalStrict(bytes, &config)

	return config, err
}