
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

import React, { Component, Fragment } from "react";
import {
  Button,
  DropdownProps,
  Form,
  InputProps,
} from "semantic-ui-react";

import "./Page.css";

interface IProps {
}

interface IState {
  type: string;
  directory: string;
  repository: string;
  branch: string;
}

export default class AddSourceForm extends Component<IProps, IState> {

  public state: IState = {
    branch: "",
    directory: "",
    repository: "",
    type: "directory",
  };

  public render() {
    const { directory, repository, branch } = this.state;
    const options = [
      { key: "directory", text: "Directory", value: "directory" },
      { key: "git", text: "Git", value: "git" },
    ];

    let typeFields: JSX.Element;

    if (this.state.type === "directory") {
      typeFields = (
        <Form.Field width="13">
          <label>Directory</label>
          <Form.Input
            name="directory"
            defaultValue={directory}
            onChange={this.handleChangeInput}
          />
        </Form.Field>
      );
    } else {
      typeFields = (
        <Fragment>
          <Form.Field width="9">
            <label>Repository</label>
            <Form.Input
              name="repository"
              defaultValue={repository}
              onChange={this.handleChangeInput}
            />
          </Form.Field>
          <Form.Field width="4">
            <label>Branch</label>
            <Form.Input
              name="branch"
              placeholder="master"
              defaultValue={branch}
              onChange={this.handleChangeInput}
            />
          </Form.Field>
        </Fragment>
      );
    }

    return (
      <Form onSubmit={this.handleSubmit}>
        <Form.Group>
          <Form.Select
            label="Type"
            options={options}
            defaultValue="directory"
            width="3"
            onChange={this.handleChangeType}
          />
          {typeFields}
        </Form.Group>
        <Button
          type="submit"
          color="teal"
          icon="add"
          content="Add"
        />
      </Form>
    );
  }

  private handleChangeType = (_: React.SyntheticEvent<HTMLElement>, { value }: DropdownProps) => {
    this.setState({ type: value as string });
  }

  private handleChangeInput = (_: React.SyntheticEvent<HTMLElement>, { name, value }: InputProps) => {
    switch (name) {
    case "directory": this.setState({ directory: value }); break;
    case "repository": this.setState({ repository: value }); break;
    case "branch": this.setState({ branch: value }); break;
    }
  }

  private handleSubmit = () => {
    console.log(this.state);
  }

}