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
	testFilename := flag.String("O", "", "A filename to output the generated test code to")
	flag.Parse()

	w := os.Stdout

	if *filename != "" {
		f, err := os.OpenFile(*filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		checkError(err)

		defer f.Close()
		defer f.Sync()

		w = f
	}

	t, err := template.New("tmpl").Parse(tmpl)
	checkError(err)

	err = t.Execute(w, strings.Split(*types, ","))
	checkError(err)

	if *testFilename != "" {
		tf, err := os.OpenFile(*testFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		checkError(err)

		defer tf.Close()
		defer tf.Sync()

		tt, err := template.New("testTmpl").Parse(testTmpl)
		checkError(err)

		err = tt.Execute(tf, strings.Split(*types, ","))
		checkError(err)
	}
}

var tmpl = `// Code generated by github.com/stratumn/groundcontrol/scripts/paginatorsgen.go, DO NOT EDIT.

package models

{{range $index, $type := .}}
// Paginate{{$type}}Slice paginates a slice of {{$type}} given query parameters.
func Paginate{{$type}}Slice(slice []{{$type}}, after, before *string, first, last *int) ({{$type}}Connection, error) {
	edgeSlice, hadMore := applyCursorsTo{{$type}}Slice(slice, after, before)
	edgeSliceLen := len(edgeSlice)

	pageInfo := PageInfo{}

	if first != nil {
		firstValue := *first
		if firstValue < 0 {
			return {{$type}}Connection{}, ErrFirstNegative
		}
		if firstValue > edgeSliceLen {
			firstValue = edgeSliceLen
		}
		if firstValue < edgeSliceLen {
			pageInfo.HasNextPage = true
		} else if before != nil {
			pageInfo.HasNextPage = hadMore
		}
		edgeSlice = edgeSlice[0:firstValue]
		edgeSliceLen = len(edgeSlice)
	} else if before != nil {
		pageInfo.HasNextPage = hadMore
	}

	if last != nil {
		lastValue := *last
		if lastValue < 0 {
			return {{$type}}Connection{}, ErrLastNegative
		}
		if lastValue > edgeSliceLen {
			lastValue = edgeSliceLen
		}
		if lastValue < edgeSliceLen {
			pageInfo.HasPreviousPage = true
		} else if after != nil {
			pageInfo.HasPreviousPage = hadMore
		}
		end := edgeSliceLen - lastValue
		edgeSlice = edgeSlice[end:]
		edgeSliceLen = len(edgeSlice)
	} else if after != nil {
		pageInfo.HasPreviousPage = hadMore
	}

	if edgeSliceLen > 0 {
		pageInfo.StartCursor = edgeSlice[0].Cursor
		pageInfo.EndCursor = edgeSlice[edgeSliceLen-1].Cursor
	}

	return {{$type}}Connection{
		Edges:    edgeSlice,
		PageInfo: pageInfo,
	}, nil
}

func applyCursorsTo{{$type}}Slice(slice []{{$type}}, after, before *string) ([]{{$type}}Edge, bool) {
	var edges []{{$type}}Edge

	hadMore := false

	if after != nil {
		index := indexOf{{$type}}InSlice(slice, *after)
		if index < 0 {
			return nil, false
		}
		hadMore = index > 0
		for _, node := range slice[index+1:] {
			edges = append(edges, {{$type}}Edge{Cursor: node.ID, Node: node})
		}
		return edges, hadMore
	}

	if before != nil {
		index := indexOf{{$type}}InSlice(slice, *before)
		if index < 0 {
			return nil, false
		}
		hadMore = index < len(slice)-1
		for _, node := range slice[:index] {
			edges = append(edges, {{$type}}Edge{Cursor: node.ID, Node: node})
		}
		return edges, hadMore
	}

	for _, node := range slice {
		edges = append(edges, {{$type}}Edge{Cursor: node.ID, Node: node})
	}
	return edges, hadMore
}

func indexOf{{$type}}InSlice(slice []{{$type}}, id string) int {
	for i := range slice {
		if slice[i].ID == id {
			return i
		}
	}

	return -1
}
{{end -}}
`

var testTmpl = `// Code generated by github.com/stratumn/groundcontrol/scripts/paginatorsgen.go, DO NOT EDIT.

package models

import (
	"fmt"
	"reflect"
	"testing"
)

{{range $index, $type := .}}
func Test_Paginate{{$type}}Slice(t *testing.T) {
	var nodes []{{$type}}
	for i := 0; i < 10; i++ {
		nodes = append(nodes, {{$type}}{
			ID:   fmt.Sprint(i),
		})
	}

	var edges []{{$type}}Edge
	for i := 0; i < 10; i++ {
		edges = append(edges, {{$type}}Edge{
			Cursor: fmt.Sprint(i),
			Node:   nodes[i],
		})
	}

	five := 5

	type args struct {
		slice  []{{$type}}
		after  *string
		before *string
		first  *int
		last   *int
	}
	tests := []struct {
		name    string
		args    args
		want    {{$type}}Connection
		wantErr bool
	}{
		{
			"nil",
			args{nil, nil, nil, nil, nil},
			{{$type}}Connection{
				Edges: nil,
				PageInfo: PageInfo{},
			},
			false,
		}, {
			"empty",
			args{[]{{$type}}{}, nil, nil, nil, nil},
			{{$type}}Connection{
				Edges: nil,
				PageInfo: PageInfo{},
			},
			false,
		}, {
			"all",
			args{nodes, nil, nil, nil, nil},
			{{$type}}Connection{
				Edges: edges,
				PageInfo: PageInfo{
					HasPreviousPage: false,
					HasNextPage:     false,
					StartCursor:     "0",
					EndCursor:       "9",
				},
			},
			false,
		}, {
			"after",
			args{nodes, &nodes[6].ID, nil, nil, nil},
			{{$type}}Connection{
				Edges: edges[7:],
				PageInfo: PageInfo{
					HasPreviousPage: true,
					HasNextPage:     false,
					StartCursor:     "7",
					EndCursor:       "9",
				},
			},
			false,
		}, {
			"before",
			args{nodes, nil, &nodes[3].ID, nil, nil},
			{{$type}}Connection{
				Edges: edges[:3],
				PageInfo: PageInfo{
					HasPreviousPage: false,
					HasNextPage:     true,
					StartCursor:     "0",
					EndCursor:       "2",
				},
			},
			false,
		}, {
			"first",
			args{nodes, nil, nil, &five, nil},
			{{$type}}Connection{
				Edges: edges[:5],
				PageInfo: PageInfo{
					HasPreviousPage: false,
					HasNextPage:     true,
					StartCursor:     "0",
					EndCursor:       "4",
				},
			},
			false,
		}, {
			"last",
			args{nodes, nil, nil, nil, &five},
			{{$type}}Connection{
				Edges: edges[5:],
				PageInfo: PageInfo{
					HasPreviousPage: true,
					HasNextPage:     false,
					StartCursor:     "5",
					EndCursor:       "9",
				},
			},
			false,
		}, {
			"after first",
			args{nodes, &nodes[2].ID, nil, &five, nil},
			{{$type}}Connection{
				Edges: edges[3:8],
				PageInfo: PageInfo{
					HasPreviousPage: true,
					HasNextPage:     true,
					StartCursor:     "3",
					EndCursor:       "7",
				},
			},
			false,
		}, {
			"after last",
			args{nodes, &nodes[2].ID, nil, nil, &five},
			{{$type}}Connection{
				Edges: edges[5:10],
				PageInfo: PageInfo{
					HasPreviousPage: true,
					HasNextPage:     false,
					StartCursor:     "5",
					EndCursor:       "9",
				},
			},
			false,
		}, {
			"before first",
			args{nodes, nil, &nodes[7].ID, &five, nil},
			{{$type}}Connection{
				Edges: edges[0:5],
				PageInfo: PageInfo{
					HasPreviousPage: false,
					HasNextPage:     true,
					StartCursor:     "0",
					EndCursor:       "4",
				},
			},
			false,
		}, {
			"before last",
			args{nodes, nil, &nodes[7].ID, nil, &five},
			{{$type}}Connection{
				Edges: edges[2:7],
				PageInfo: PageInfo{
					HasPreviousPage: true,
					HasNextPage:     true,
					StartCursor:     "2",
					EndCursor:       "6",
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Paginate{{$type}}Slice(tt.args.slice, tt.args.after, tt.args.before, tt.args.first, tt.args.last)
			if (err != nil) != tt.wantErr {
				t.Errorf("Paginate{{$type}}Slice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Paginate{{$type}}Slice() = %v, want %v", got, tt.want)
			}
		})
	}
}
{{end -}}
`
