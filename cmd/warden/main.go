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
	"os"
	"strconv"
	"time"

	warden "github.com/saashqdev/kubeworkz/cmd/warden/app"
	"github.com/saashqdev/kubeworkz/cmd/warden/app/options"

	"github.com/urfave/cli/v2"
)

var version = "1.0.0"

func main() {
	app := cli.NewApp()
	app.Name = "Warden"
	app.Usage = "Warden will keep watching in kube member cluster"
	app.Version = version
	app.Compiled = time.Now()
	app.Copyright = "(c) " + strconv.Itoa(time.Now().Year()) + " Kubeworkz"

	app.Flags = options.Flags
	app.Action = warden.Start

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
