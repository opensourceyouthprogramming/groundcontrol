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

import graphql from "babel-plugin-relay/macro";
import { Link } from "found";
import React, { Component } from "react";
import {
  Table,
 } from "semantic-ui-react";

import { createFragmentContainer } from "react-relay";

import { JobTableRow_item } from "./__generated__/JobTableRow_item.graphql";

import Moment from "react-moment";
import RepositoryShortName from "./RepositoryShortName";

const dateFormat = "L LTS";

interface IProps {
  item: JobTableRow_item;
}

export class JobTableRow extends Component<IProps> {

  public render() {
    const item = this.props.item;

    return (
      <Table.Row>
        <Table.Cell>{item.name}</Table.Cell>
        <Table.Cell>
          <Link to={`/workspaces/${item.project.workspace.slug}`}>
            {item.project.workspace.name}
          </Link>
        </Table.Cell>
        <Table.Cell>
          <RepositoryShortName repository={item.project.repository} />
        </Table.Cell>
        <Table.Cell>{item.project.branch}</Table.Cell>
        <Table.Cell>
          <Moment format={dateFormat}>{item.createdAt}</Moment>
        </Table.Cell>
        <Table.Cell>
          <Moment format={dateFormat}>{item.updatedAt}</Moment>
        </Table.Cell>
        <Table.Cell
          positive={item.status === "DONE"}
          warning={item.status === "RUNNING"}
          negative={item.status === "FAILED"}
        >
          {item.status}
        </Table.Cell>
      </Table.Row>
    );
  }

}

export default createFragmentContainer(JobTableRow, graphql`
  fragment JobTableRow_item on Job {
    name
    status
    createdAt
    updatedAt
    project {
      repository
      branch
      workspace {
        slug
        name
      }
    }
  }`,
);