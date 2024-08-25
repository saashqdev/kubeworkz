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

package apiserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	_ "github.com/saashqdev/kubeworkz/docs"
	_ "github.com/saashqdev/kubeworkz/pkg/apis/user/v1"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/authorization"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/cluster"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/healthz"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/k8s"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/key"
	resourcemanage "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/handle"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/scout"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/user"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/yamldeploy"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/middlewares"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	"github.com/saashqdev/kubeworkz/pkg/utils/constants"
	_ "github.com/saashqdev/kubeworkz/pkg/utils/errcode"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// APIServer aggregates all kube apis
type APIServer struct {
	*Config

	Server       *http.Server
	SimpleServer *http.Server
}

// @title Swagger Kubeworkz API
// @version 1.0
// @description This is Kubeworkz api documentation.
// registerKubeAPI register apis for kube api server
func registerKubeAPI(cfg *Config) http.Handler {
	router := gin.New()

	// register apis do not need middlewares
	apisOutsideMiddlewares(router)

	// set middlewares for apis below
	middlewares.SetUpMiddlewares(router, cfg.Gi18nManagers)

	// clusters apis handler
	cluster.NewHandler().AddApisTo(router)

	// authZ apis handler
	authorization.NewHandler().AddApisTo(router)

	router.POST(constants.ApiPathRoot+"/login", user.Login)
	router.GET(constants.ApiPathRoot+"/oauth/redirect", user.GitHubLogin)

	userManage := router.Group(constants.ApiPathRoot + "/user")
	{
		userManage.POST("/", user.CreateUser)
		userManage.PUT("/:username", user.UpdateUser)
		userManage.GET("", user.ListUsers)
		userManage.GET("/template", user.DownloadTemplate)
		userManage.POST("/users", user.BatchCreateUser)
		userManage.GET("/kubeconfigs", user.GetKubeConfig)
		userManage.GET("/members", user.GetMembersByNS)
		userManage.GET("/valid/:username", user.CheckUserValid)
		userManage.PUT("/pwd", user.UpdatePwd)
	}

	keyManage := router.Group(constants.ApiPathRoot + "/key")
	{
		keyManage.GET("/token", key.GetTokenByKey)
		keyManage.POST("/create", key.CreateKey)
		keyManage.DELETE("", key.DeleteKey)
		keyManage.GET("", key.ListKey)
	}

	k8sApiProxy := router.Group(constants.ApiPathRoot + "/proxy")
	{
		proxyHandler := resourcemanage.NewProxyHandler(cfg.EnableVersionConversion)
		k8sApiProxy.Any("/clusters/:cluster/*url", proxyHandler.ProxyHandle)
	}

	k8sApiExtend := router.Group(constants.ApiPathRoot + "/extend")
	{
		extendHandler := resourcemanage.NewExtendHandler(cfg.NginxNamespace, cfg.NginxTcpServiceConfigMap, cfg.NginxUdpServiceConfigMap)
		yamlDeployHandler := yamldeploy.NewHandler(cfg.Gi18nManagers)

		k8sApiExtend.GET("/feature-config", resourcemanage.GetFeatureConfig)
		k8sApiExtend.Any("/clusters/:cluster/resources/:resourceType", extendHandler.ExtendHandle)
		k8sApiExtend.Any("/clusters/:cluster/namespaces/:namespace/:resourceType/:resourceName", extendHandler.ExtendHandle)
		k8sApiExtend.Any("/clusters/:cluster/namespaces/:namespace/:resourceType", extendHandler.ExtendHandle)
		k8sApiExtend.GET("/clusters/:cluster/namespaces/:namespace/logs/:resourceName", resourcemanage.GetPodContainerLog)
		k8sApiExtend.GET("/clusters/:cluster/namespaces/:namespace/proxy/logs/:resourceName", resourcemanage.GetProxyPodContainerLog)
		k8sApiExtend.POST("/clusters/:cluster/yaml/deploy", yamlDeployHandler.Deploy)
		k8sApiExtend.GET("/ingressDomainSuffix", resourcemanage.IngressDomainSuffix)
	}

	return router
}

func NewAPIServerWithOpts(opts *Config) *APIServer {
	router := registerKubeAPI(opts)

	s := &APIServer{
		Server: &http.Server{
			Handler: router,
			Addr:    fmt.Sprintf("%s:%d", opts.BindAddr, opts.InsecurePort),
		},
		Config: opts,
	}

	if opts.SecurePort != 0 {
		s.Server.Addr = fmt.Sprintf("%s:%d", s.Config.BindAddr, s.Config.SecurePort)
	}

	return withSimpleServer(s)
}

func withSimpleServer(s *APIServer) *APIServer {
	router := gin.New()
	router.Use(gin.Recovery())

	if s.SwagEnable {
		// The url pointing to API definition
		url := ginSwagger.URL("/swagger/doc.json")
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	}

	router.GET("/healthz", healthz.HealthyCheck)
	router.PUT("/log/level", gin.WrapH(http.HandlerFunc(clog.GetAtomicLevel().ServeHTTP)))

	s.SimpleServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.Config.BindAddr, s.Config.GenericPort),
		Handler: router,
	}

	return s
}

func apisOutsideMiddlewares(root *gin.Engine) {
	scout.AddApisTo(root)

	root.GET(constants.ApiPathRoot+"/extend/configmap/:configmap", resourcemanage.GetConfigMap)
	root.Any(constants.ApiK8sProxyPath+"/*path", k8s.NewHandler().LocalClusterProxy)
}

func (s *APIServer) Initialize() error {
	return nil
}

func (s *APIServer) Run(stop <-chan struct{}) {
	go func() {
		var err error

		if s.Config.SecurePort != 0 {
			err = s.Server.ListenAndServeTLS(s.Config.TlsCert, s.Config.TlsKey)
		} else {
			err = s.Server.ListenAndServe()
		}

		if err != nil {
			clog.Fatal("kube apiserver start err: %v", err)
		}
	}()

	go func() {
		err := s.SimpleServer.ListenAndServe()
		if err != nil {
			clog.Fatal("kube simple server start err: %v", err)
		}
	}()

	clog.Info("kube apiserver listen in %s", s.Server.Addr)
	clog.Info("kube simple server listen in %v", s.SimpleServer.Addr)

	<-stop

	clog.Info("shutting down kube apiserver and simple server")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Server.Shutdown(ctx); err != nil {
		clog.Fatal("kube apiserver forced to shutdown: %v", err)
	}

	if err := s.SimpleServer.Shutdown(ctx); err != nil {
		clog.Fatal("kube simple server forced to shutdown: %v", err)
	}

	clog.Info("kube apiserver and simple server exiting")
}
