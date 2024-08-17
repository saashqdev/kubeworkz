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

// api-server flags
func init() {
	Flags = append(Flags, []cli.Flag{
		// Server flags
		&cli.StringFlag{
			Name:        "bind-addr",
			Value:       "0.0.0.0",
			Destination: &KubeOpts.APIServerOpts.BindAddr,
		},
		&cli.IntFlag{
			Name:        "insecure-port",
			Destination: &KubeOpts.APIServerOpts.InsecurePort,
		},
		&cli.IntFlag{
			Name:        "secure-port",
			Value:       7443,
			Destination: &KubeOpts.APIServerOpts.SecurePort,
		},
		&cli.IntFlag{
			Name:        "generic-port",
			Value:       7777,
			Destination: &KubeOpts.APIServerOpts.GenericPort,
		},
		&cli.BoolFlag{
			Name:        "enable-swag",
			Value:       false,
			Destination: &KubeOpts.APIServerOpts.SwagEnable,
		},
		&cli.StringFlag{
			Name:        "tls-cert",
			Destination: &KubeOpts.APIServerOpts.TlsCert,
		},
		&cli.StringFlag{
			Name:        "tls-key",
			Destination: &KubeOpts.APIServerOpts.TlsKey,
		},
		&cli.StringFlag{
			Name:        "ca-cert",
			Destination: &KubeOpts.APIServerOpts.CaCert,
		},
		&cli.StringFlag{
			Name:        "ca-key",
			Destination: &KubeOpts.APIServerOpts.CaKey,
		},
		// todo(weilaaa): move this flag to suitable place
		&cli.BoolFlag{
			Name:        "enable-version-conversion",
			Value:       false,
			Destination: &KubeOpts.APIServerOpts.EnableVersionConversion,
		},
		&cli.StringFlag{
			Name:        "ingress-nginx-namespace",
			Value:       "ingress-nginx",
			Destination: &KubeOpts.APIServerOpts.NginxNamespace,
		},
		&cli.StringFlag{
			Name:        "ingress-nginx-tcp-configmap",
			Value:       "tcp-services",
			Destination: &KubeOpts.APIServerOpts.NginxTcpServiceConfigMap,
		},
		&cli.StringFlag{
			Name:        "ingress-nginx-udp-configmap",
			Value:       "udp-services",
			Destination: &KubeOpts.APIServerOpts.NginxUdpServiceConfigMap,
		},
	}...)
}
