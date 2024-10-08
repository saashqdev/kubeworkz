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

package authproxy

import (
	"context"

	"net/http"
	"net/url"
	"strings"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/saashqdev/kubeworkz/pkg/authentication/authenticators"
	"github.com/saashqdev/kubeworkz/pkg/authentication/authenticators/jwt"
	"github.com/saashqdev/kubeworkz/pkg/authentication/authenticators/token"
	"github.com/saashqdev/kubeworkz/pkg/belongs"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	"github.com/saashqdev/kubeworkz/pkg/multicluster/client"
	"github.com/saashqdev/kubeworkz/pkg/utils/constants"
	requestutil "github.com/saashqdev/kubeworkz/pkg/utils/request"
	"github.com/saashqdev/kubeworkz/pkg/warden/server/authproxy/proxy"
)

// Handler forwards all the requests to specified k8s-apiserver
// after pass previous authentication
type Handler struct {
	// authMgr has the way to operator jwt token
	authMgr authenticators.AuthNManager

	// cfg holds current cluster info
	// cfg *rest.Config

	cli client.Client

	// proxy do real proxy action with any inbound stream
	proxy *proxy.UpgradeAwareHandler
}

func NewHandler(localClusterKubeConfig string) (*Handler, error) {
	// get cluster info from rest config
	restConfig, err := clientcmd.BuildConfigFromFlags("", localClusterKubeConfig)
	if err != nil {
		return nil, err
	}
	h := &Handler{}
	err = h.SetHandlerClientByRestConfig(restConfig)
	if err != nil {
		return nil, err
	}
	err = h.SetHandlerTS(restConfig)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (h *Handler) SetHandlerTS(restConfig *rest.Config) error {
	h.authMgr = jwt.GetAuthJwtImpl()

	host := restConfig.Host
	if !strings.HasSuffix(host, "/") {
		host = host + "/"
	}
	target, err := url.Parse(host)
	if err != nil {
		return err
	}

	responder := &responder{}
	ts, err := rest.TransportFor(restConfig)
	if err != nil {
		return err
	}

	upgradeTransport, err := makeUpgradeTransport(restConfig, 30*time.Second)
	if err != nil {
		return err
	}

	p := proxy.NewUpgradeAwareHandler(target, ts, false, false, responder)
	p.UpgradeTransport = upgradeTransport
	p.UseRequestLocation = true

	h.proxy = p

	return nil
}

func (h *Handler) SetHandlerClient(cli client.Client) {
	h.cli = cli
}

func (h *Handler) SetHandlerClientByRestConfig(restConfig *rest.Config) error {
	cli, err := client.NewClientFor(context.Background(), restConfig)
	if err != nil {
		return err
	}
	h.cli = cli
	return nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// parse token transfer to user info
	userInfo, err := token.GetUserFromReq(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	clog.Debug("user(%v) access to %v with verb(%v)", userInfo.Username, r.URL.Path, r.Method)

	allowed, err := belongs.RelationshipDetermine(context.Background(), h.cli, r.URL.Path, userInfo.Username)
	if err != nil {
		clog.Warn(err.Error())
	} else if !allowed {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = requestutil.AddFieldManager(r, userInfo.Username)
	if err != nil {
		clog.Error("fail to add fieldManager due to %s", err)
	}

	// impersonate given user to access k8s-apiserver
	r.Header.Set(constants.ImpersonateUserKey, userInfo.Username)
	r.Header.Del(constants.AuthorizationHeader)
	h.proxy.ServeHTTP(w, r)
}
