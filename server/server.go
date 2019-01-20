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

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/stratumn/groundcontrol"
)

const defaultPort = "8080"

var ui http.FileSystem

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	args := os.Args
	if len(args) > 2 {
		fmt.Printf("usage: %s [file]\n", args[0])
		os.Exit(1)
	}

	filename, err := filepath.Abs("groundcontrol.yml")
	checkError(err)

	if len(args) > 1 {
		filename, err = filepath.Abs(args[1])
		checkError(err)
	}

	resolver := groundcontrol.Resolver{}
	err = groundcontrol.LoadYAML(filename, &groundcontrol.Viewer)
	checkError(err)

	config := groundcontrol.Config{
		Resolvers: &resolver,
	}

	c := cors.New(cors.Options{
		AllowCredentials: true,
		Debug:            false,
	})

	router := chi.NewRouter()
	router.Use(c.Handler)

	if ui != nil {
		fileServer := http.FileServer(ui)
		router.Handle("/", fileServer)
		router.NotFound(fileServer.ServeHTTP)
	} else {
		router.Handle("/", handler.Playground("GraphQL playground", "/query"))
	}

	router.Handle("/query", handler.GraphQL(
		groundcontrol.NewExecutableSchema(config),
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		}),
	))

	if ui != nil {
		log.Printf("connect to http://localhost:%s/ for web interface", port)
	} else {
		log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	}

	log.Fatal(http.ListenAndServe(":"+port, router))
}
