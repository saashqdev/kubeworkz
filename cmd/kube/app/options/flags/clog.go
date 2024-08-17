/*
Copyright 2024 KubeWorkz Authors

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

package flags

import "github.com/urfave/cli/v2"

// clog flags
func init() {
	Flags = append(Flags, []cli.Flag{
		// rotate flags
		&cli.StringFlag{
			Name:        "log-file",
			Value:       "/etc/logs/kube.log",
			Destination: &KubeOpts.KubeLoggerOpts.LogFile,
		},
		&cli.IntFlag{
			Name:        "max-size",
			Value:       1000,
			Destination: &KubeOpts.KubeLoggerOpts.MaxSize,
		},
		&cli.IntFlag{
			Name:        "max-backups",
			Value:       7,
			Destination: &KubeOpts.KubeLoggerOpts.MaxBackups,
		},
		&cli.IntFlag{
			Name:        "max-age",
			Value:       1,
			Destination: &KubeOpts.KubeLoggerOpts.MaxAge,
		},
		&cli.BoolFlag{
			Name:        "compress",
			Value:       true,
			Destination: &KubeOpts.KubeLoggerOpts.Compress,
		},

		// logger flags
		&cli.StringFlag{
			Name:        "log-level",
			Value:       "info",
			Destination: &KubeOpts.KubeLoggerOpts.LogLevel,
		},
		&cli.BoolFlag{
			Name:        "json-encode",
			Value:       false,
			Destination: &KubeOpts.KubeLoggerOpts.JsonEncode,
		},
		&cli.StringFlag{
			Name:        "stacktrace-level",
			Value:       "error",
			Destination: &KubeOpts.KubeLoggerOpts.StacktraceLevel,
		},
	}...)
}
