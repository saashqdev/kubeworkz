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

package options

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"github.com/saashqdev/kubeworkz/pkg/apiserver"
	"github.com/saashqdev/kubeworkz/pkg/authentication"
	"github.com/saashqdev/kubeworkz/pkg/clients"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	"github.com/saashqdev/kubeworkz/pkg/ctrlmgr"
	"github.com/saashqdev/kubeworkz/pkg/kube"
)

const (
	defaultConfiguration = "kubeworkz"
	defaultConfigPath    = "/etc/kubeworkz"
)

type KubeOptions struct {
	GenericKubeOpts *kube.Config
	APIServerOpts   *apiserver.Config
	CtrlMgrOpts     *ctrlmgr.Config
	ClientMgrOpts   *clients.Config
	KubeLoggerOpts  *clog.Config
	AuthMgrOpts     *authentication.Config
}

func NewKubeOptions() *KubeOptions {
	kubeOpts := &KubeOptions{
		GenericKubeOpts: &kube.Config{},
		APIServerOpts:   &apiserver.Config{},
		CtrlMgrOpts:     &ctrlmgr.Config{},
		ClientMgrOpts:   &clients.Config{},
		KubeLoggerOpts:  &clog.Config{},
		AuthMgrOpts:     &authentication.Config{},
	}

	return kubeOpts
}

func LoadConfigFromDisk() (*KubeOptions, error) {
	viper.SetConfigName(defaultConfiguration)
	viper.AddConfigPath(defaultConfigPath)
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError *viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return nil, err
		} else {
			return nil, fmt.Errorf("error parsing configuration file %s", err)
		}
	}

	conf := NewKubeOptions()

	if err := viper.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

// Validate verify options for every component
// todo(weilaaa): complete it
func (s *KubeOptions) Validate() []error {
	var errs []error

	errs = append(errs, s.APIServerOpts.Validate()...)
	errs = append(errs, s.ClientMgrOpts.Validate()...)
	errs = append(errs, s.CtrlMgrOpts.Validate()...)

	return errs
}

func (s *KubeOptions) NewKube() *kube.Kube {
	return &kube.Kube{}
}
