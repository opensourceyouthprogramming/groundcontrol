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

//+build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	types := flag.String("t", "", "A comma separated list of types")
	filename := flag.String("o", "", "A filename to output the generated code to")
	flag.Parse()

	w := os.Stdout

	if *filename != "" {
		f, err := os.OpenFile(*filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		checkError(err)

		defer f.Close()
		defer f.Sync()

		w = f
	}

	t, err := template.New("tmpl").Parse(strings.TrimSpace(tmpl))
	checkError(err)

	err = t.Execute(w, strings.Split(*types, ","))
	checkError(err)
}

var tmpl = `
// Code generated by groundcontrol/scripts/nodesgen.go, DO NOT EDIT.

package models

import "groundcontrol/relay"

// Node types.
const (
{{- range $index, $type := .}}
	NodeType{{$type}} = "{{$type}}"
{{- end}}
)

{{range $index, $type := .}}
// GetID returns the unique ID of the node.
func (n {{$type}}) GetID() string {
	return n.ID
}
{{end}}

{{range $index, $type := .}}
// Store{{$type}} stores a {{$type}}.
func (n *NodeManager) Store{{$type}}(node {{$type}}) error {
	identifiers, err := relay.DecodeID(node.GetID())
	if err != nil {
		return err
	}
	if identifiers[0] != NodeType{{$type}} {
		return ErrType
	}

	n.store.Store(node.GetID(), node)

	return nil
}

// MustStore{{$type}} stores a {{$type}} or panics on failure.
func (n *NodeManager) MustStore{{$type}}(node {{$type}}) {
	if err := n.Store{{$type}}(node); err != nil {
		panic(err)
	}
}

// Load{{$type}} loads a {{$type}}.
func (n *NodeManager) Load{{$type}}(id string) ({{$type}}, error) {
	identifiers, err := relay.DecodeID(id)
	if err != nil {
		return {{$type}}{}, err
	}
	if identifiers[0] != NodeType{{$type}} {
		return {{$type}}{}, ErrType
	}
	node, ok := n.store.Load(id)
	if !ok {
		return {{$type}}{}, ErrNotFound
	}

	return node.({{$type}}), nil
}

// MustLoad{{$type}} loads a {{$type}} or panics on failure.
func (n *NodeManager) MustLoad{{$type}}(id string) {{$type}} {
	node, err := n.Load{{$type}}(id)
	if err != nil {
		panic(err)
	}

	return node
}

// Delete{{$type}} deletes a {{$type}}.
// If the node doesn't exist it's a NOP.
func (n *NodeManager) Delete{{$type}}(id string) error {
	identifiers, err := relay.DecodeID(id)
	if err != nil {
		return err
	}
	if identifiers[0] != NodeType{{$type}} {
		return ErrType
	}

	n.store.Delete(id)
	return nil
}

// MustDelete{{$type}} deletes a {{$type}} or panics on failure.
// If the node doesn't exist it's a NOP.
func (n *NodeManager) MustDelete{{$type}}(id string) {
	err := n.Delete{{$type}}(id)
	if err != nil {
		panic(err)
	}
}

// Lock{{$type}} loads a {{$type}} and locks it until the callback returns.
func (n *NodeManager) Lock{{$type}}(id string, fn func({{$type}})) error {
	n.Lock(id)

	node, err := n.Load{{$type}}(id)
	if err != nil {
		return err
	}

	fn(node)
	n.Unlock(id)

	return nil
}

// Lock{{$type}}E is like Lock{{$type}}, but the callback can return an error.
func (n *NodeManager) Lock{{$type}}E(id string, fn func({{$type}}) error) error {
	n.Lock(id)

	node, err := n.Load{{$type}}(id)
	if err != nil {
		return err
	}

	err = fn(node)
	n.Unlock(id)

	return err
}

// MustLock{{$type}} loads a {{$type}} or panics on error and locks it until the callback returns.
func (n *NodeManager) MustLock{{$type}}(id string, fn func({{$type}})) {
	n.Lock(id)

	node, err := n.Load{{$type}}(id)
	if err != nil {
		panic(err)
	}

	fn(node)
	n.Unlock(id)
}

// MustLock{{$type}}E is like MustLock{{$type}}, but the callback can return an error.
func (n *NodeManager) MustLock{{$type}}E(id string, fn func({{$type}}) error) error {
	n.Lock(id)

	node, err := n.Load{{$type}}(id)
	if err != nil {
		panic(err)
	}

	err = fn(node)
	n.Unlock(id)

	return err
}

// LockOrNew{{$type}} loads or initializes a {{$type}} and locks it until the callback returns.
func (n *NodeManager) LockOrNew{{$type}}(id string, fn func({{$type}})) error {
	n.Lock(id)

	node, err := n.Load{{$type}}(id)
	if err == ErrNotFound {
		node = {{$type}}{
			ID: id,
		}
	} else if err != nil {
		return err
	}

	fn(node)
	n.Unlock(id)

	return nil
}

// LockOrNew{{$type}}E is like LockOrNew{{$type}}, but the callback can return an error.
func (n *NodeManager) LockOrNew{{$type}}E(id string, fn func({{$type}}) error) error {
	n.Lock(id)

	node, err := n.Load{{$type}}(id)
	if err == ErrNotFound {
		node = {{$type}}{
			ID: id,
		}
	} else if err != nil {
		return err
	}

	err = fn(node)
	n.Unlock(id)

	return err
}

// MustLockOrNew{{$type}} loads or initializes a {{$type}} or panics on error and locks it until the callback returns.
func (n *NodeManager) MustLockOrNew{{$type}}(id string, fn func({{$type}})) {
	n.Lock(id)

	node, err := n.Load{{$type}}(id)
	if err == ErrNotFound {
		node = {{$type}}{
			ID: id,
		}
	} else if err != nil {
		panic(err)
	}

	fn(node)
	n.Unlock(id)
}

// MustLockOrNew{{$type}}E is like MustLockOrNew{{$type}}, but the callback can return an error.
func (n *NodeManager) MustLockOrNew{{$type}}E(id string, fn func({{$type}}) error) error {
	n.Lock(id)

	node, err := n.Load{{$type}}(id)
	if err == ErrNotFound {
		node = {{$type}}{
			ID: id,
		}
	} else if err != nil {
		panic(err)
	}

	err = fn(node)
	n.Unlock(id)

	return err
}

{{end}}
`
