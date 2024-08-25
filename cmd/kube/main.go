/*
Copyright 2024 Kubeworkz Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/saashqdev/kubeworkz/cmd/kube/app/options/flags"

	kube "github.com/saashqdev/kubeworkz/cmd/kube/app"

	"github.com/urfave/cli/v2"
)

var version = "1.0.0"

func main() {
	rand.Seed(time.Now().UnixNano())

	app := cli.NewApp()
	app.Name = "Kubeworkz"
	app.Usage = "KubWorkz K8s for the rest of us"
	app.Version = version
	app.Compiled = time.Now()
	app.Copyright = "(c) " + strconv.Itoa(time.Now().Year()) + " Kubeworkz"

	app.Flags = flags.Flags
	app.Before = kube.Before
	app.Action = kube.Start

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
