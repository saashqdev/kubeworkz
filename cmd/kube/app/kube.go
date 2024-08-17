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

package app

import (
	"github.com/saashqdev/kubeworkz/cmd/kube/app/options"
	"github.com/saashqdev/kubeworkz/cmd/kube/app/options/flags"
	"github.com/saashqdev/kubeworkz/pkg/apiserver"
	_ "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/register"
	"github.com/saashqdev/kubeworkz/pkg/clients"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	"github.com/saashqdev/kubeworkz/pkg/ctrlmgr"
	"github.com/saashqdev/kubeworkz/pkg/kube"
	"github.com/saashqdev/kubeworkz/pkg/utils/international"
	"github.com/urfave/cli/v2"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	"k8s.io/sample-controller/pkg/signals"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	Before = func(c *cli.Context) error {
		if !c.Bool("useConfigFile") {
			return nil
		}

		var err error
		flags.KubeOpts, err = options.LoadConfigFromDisk()
		if err != nil {
			return err
		}

		return nil
	}

	Start = func(c *cli.Context) error {
		if errs := flags.KubeOpts.Validate(); len(errs) > 0 {
			return utilerrors.NewAggregate(errs)
		}

		run(flags.KubeOpts, signals.SetupSignalHandler())

		return nil
	}
)

func run(s *options.KubeOptions, stop <-chan struct{}) {
	// init kube logger first
	clog.InitKubeLoggerWithOpts(flags.KubeOpts.KubeLoggerOpts)

	// init setting klog level
	var klogLevel klog.Level
	if err := klogLevel.Set(flags.KubeOpts.GenericKubeOpts.KlogLevel); err != nil {
		clog.Fatal("klog level set failed: %v", err)
	}

	log.SetLogger(klog.NewKlogr())
	// initialize kube client set
	clients.InitKubeClientSetWithOpts(s.ClientMgrOpts)

	// initialize language managers
	m, err := international.InitGi18nManagers()
	if err != nil {
		clog.Fatal("kube initialized gi18n managers failed: %v", err)
	}
	s.APIServerOpts.Gi18nManagers = m

	c := kube.New(s.GenericKubeOpts)
	c.IntegrateWith("kube-controller-manager", ctrlmgr.NewCtrlMgrWithOpts(s.CtrlMgrOpts))
	c.IntegrateWith("kube-apiserver", apiserver.NewAPIServerWithOpts(s.APIServerOpts))

	err = c.Initialize()
	if err != nil {
		clog.Fatal("kube initialized failed: %v", err)
	}

	c.Run(stop)
}
