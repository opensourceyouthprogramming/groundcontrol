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
import { Router } from "found";
import React, { Component } from "react";
import { createPaginationContainer, RelayPaginationProp } from "react-relay";
import { Disposable } from "relay-runtime";
import { Button } from "semantic-ui-react";

import { LogEntryListPage_system } from "./__generated__/LogEntryListPage_system.graphql";

import LogEntryFilter from "../components/LogEntryFilter";
import LogEntryTable from "../components/LogEntryTable";
import Page from "../components/Page";

import { subscribe } from "../subscriptions/logEntryAdded";

interface IProps {
  relay: RelayPaginationProp;
  router: Router;
  system: LogEntryListPage_system;
  params: {
    filters: string | undefined;
  };
}

export class LogEntryListPage extends Component<IProps> {

  private disposables: Disposable[] = [];

  public render() {
    const items = this.props.system.logEntries.edges.map(({ node }) => node);
    const filters = this.props.params.filters === undefined ? undefined :
      this.props.params.filters.split(",");

    return (
      <Page
        header="Logs"
        subheader="Logs are short messages emitted after events of various levels."
        icon="book"
      >
        <LogEntryFilter
          filters={filters}
          onChange={this.handleFiltersChange}
        />
        <LogEntryTable items={items} />
        <Button
          disabled={!this.props.relay.hasMore() || this.props.relay.isLoading()}
          loading={this.props.relay.isLoading()}
          color="grey"
          onClick={this.handleLoadMore}
        >
          Load More
        </Button>
      </Page>
    );
  }

  public componentDidMount() {
    this.disposables.push(subscribe(this.props.relay.environment));
  }

  public componentWillUnmount() {
    for (const disposable of this.disposables) {
      disposable.dispose();
    }

    this.disposables = [];
  }

  private handleFiltersChange = (filters: string[]) => {
    if (filters.length < 1 || filters.length > 3) {
      return this.props.router.replace("/logs");
    }

    this.props.router.replace(`/logs/${filters.join(",")}`);
  }

  private handleLoadMore = () => {
    this.props.relay.loadMore(
      10,
      (err) => {
        if (err) {
          console.log(err);
        }

        // Make sure load more button updates.
        this.forceUpdate();
      },
    );
  }

}

export default createPaginationContainer(
  LogEntryListPage,
  graphql`
    fragment LogEntryListPage_system on System
      @argumentDefinitions(
        count: {type: "Int", defaultValue: 10},
        cursor: {type: "String"},
        level: { type: "[LogLevel!]", defaultValue: null },
      ) {
      logEntries(
       first: $count,
       after: $cursor,
       level: $level,
      )
        @connection(
          key: "LogEntryListPage_logEntries",
          filters: ["level"],
        ) {
        edges {
          node {
            ...LogEntryTable_items
            id
          }
        }
      }
    }`,
  {
    direction: "forward",
    getConnectionFromProps: (props) => props.system && props.system.logEntries,
    getVariables: (_, {count, cursor}, fragmentVariables) => ({
      count,
      cursor,
      level: fragmentVariables.level,
    }),
    query: graphql`
      query LogEntryListPagePaginationQuery(
        $count: Int!,
        $cursor: String,
        $level: [LogLevel!],
      ) {
        system {
          ...LogEntryListPage_system @arguments(
            count: $count,
            cursor: $cursor,
            level: $level,
          )
        }
      }
    `,
  },
);