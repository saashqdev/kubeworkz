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

package resourcemanage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/saashqdev/kubeworkz/pkg/belongs"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	"github.com/saashqdev/kubeworkz/pkg/conversion"
	"github.com/saashqdev/kubeworkz/pkg/multicluster"
	"github.com/saashqdev/kubeworkz/pkg/utils/constants"
	"github.com/saashqdev/kubeworkz/pkg/utils/errcode"
	"github.com/saashqdev/kubeworkz/pkg/utils/filter"
	"github.com/saashqdev/kubeworkz/pkg/utils/page"
	requestutil "github.com/saashqdev/kubeworkz/pkg/utils/request"
	"github.com/saashqdev/kubeworkz/pkg/utils/response"
	"github.com/saashqdev/kubeworkz/pkg/utils/selector"
	"github.com/saashqdev/kubeworkz/pkg/utils/sort"
)

type ProxyHandler struct {
	// enableConvert means proxy handler will convert resources
	enableConvert bool
	// converter the version converter for doing resources convert
	converter conversion.MultiVersionConverter
}

func NewProxyHandler(enableConvert bool) *ProxyHandler {
	return &ProxyHandler{
		enableConvert: enableConvert,
		converter:     multicluster.NewDefaultMultiVersionConverter(multicluster.Interface()),
	}
}

// tryVersionConvert try to convert url and request body by given target cluster
func (h *ProxyHandler) tryVersionConvert(cluster, url string, req *http.Request) (bool, []byte, string, error) {
	if !h.enableConvert {
		return false, nil, "", nil
	}

	_, isNamespaced, gvr, err := conversion.ParseURL(url)
	if err != nil {
		return false, nil, "", err
	}
	converter, err := h.converter.GetVersionConvert(cluster)
	if err != nil {
		return false, nil, "", err
	}
	greetBack, _, recommendVersion, err := converter.GvrGreeting(gvr)
	if err != nil {
		// we just record error and pass through anyway
		clog.Warn(err.Error())
	}
	if greetBack != conversion.IsNeedConvert {
		// pass through anyway if not need convert
		clog.Info("%v greet cluster %v is %v, pass through", gvr.String(), cluster, greetBack)
		return false, nil, "", nil
	}
	if recommendVersion == nil {
		return false, nil, "", nil
	}

	// convert url according to specified gvr at first
	convertedUrl, err := conversion.ConvertURL(url, &schema.GroupVersionResource{Group: recommendVersion.Group, Version: recommendVersion.Version, Resource: gvr.Resource})
	if err != nil {
		return false, nil, "", err
	}

	// we do not need convert body if request not create and update
	if req.Method != http.MethodPost && req.Method != http.MethodPut {
		return true, nil, convertedUrl, nil
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		return false, nil, "", err
	}
	// decode data into internal version of object
	raw, rawGvr, err := converter.Decode(data, nil, nil)
	if err != nil {
		return false, nil, "", err
	}
	if rawGvr.GroupVersion().String() != gvr.GroupVersion().String() {
		return false, nil, "", fmt.Errorf("gv parse failed with pair(%v~%v)", rawGvr.GroupVersion().String(), gvr.GroupVersion().String())
	}
	// covert internal version object int recommend version object
	out, err := converter.Convert(raw, nil, recommendVersion.GroupVersion())
	if err != nil {
		return false, nil, "", err
	}
	// encode concerted object
	convertedObj, err := converter.Encode(out, recommendVersion.GroupVersion())
	if err != nil {
		return false, nil, "", err
	}

	objMeta, err := meta.Accessor(out)
	if err != nil {
		return false, nil, "", err
	}

	if isNamespaced {
		clog.Info("resource (%v/%v) converted with (%v~%v) when visit cluster %v", objMeta.GetNamespace(), objMeta.GetName(), gvr.String(), recommendVersion.GroupVersion().WithResource(gvr.Resource), cluster)
	} else {
		clog.Info("resource (%v) converted with (%v~%v) when visit cluster %v", objMeta.GetName(), gvr.String(), recommendVersion.GroupVersion().WithResource(gvr.Resource), cluster)
	}

	return true, convertedObj, convertedUrl, nil
}

// ProxyHandle proxy all requests access to k8s, request uri format like below
// api/v1/kube/proxy/clusters/{cluster}/{k8s_url}
func (h *ProxyHandler) ProxyHandle(c *gin.Context) {
	// http request params
	cluster := c.Param("cluster")
	proxyUrl := c.Param("url")
	username := c.GetString(constants.UserName)
	if len(username) == 0 {
		clog.Warn("username is empty")
	}
	condition := parseQueryParams(c)
	converterContext := filter.ConverterContext{}
	c.Request.Header.Set(constants.ImpersonateUserKey, username)
	internalCluster, err := multicluster.Interface().Get(cluster)
	if err != nil {
		response.FailReturn(c, errcode.BadRequest(err))
		return
	}
	transport, err := multicluster.Interface().GetTransport(cluster)
	if err != nil {
		response.FailReturn(c, errcode.BadRequest(err))
		return
	}
	_, _, gvr, err := conversion.ParseURL(proxyUrl)
	if err != nil {
		response.FailReturn(c, errcode.BadRequest(err))
		return
	}

	needConvert, convertedObj, convertedUrl, err := h.tryVersionConvert(cluster, proxyUrl, c.Request)
	if err != nil {
		response.FailReturn(c, errcode.BadRequest(err))
		return
	}

	allowed, err := belongs.RelationshipDetermine(context.Background(), internalCluster.Client, proxyUrl, username)
	if err != nil {
		clog.Warn(err.Error())
	} else if !allowed {
		response.FailReturn(c, errcode.ForbiddenErr)
		return
	}

	// create director
	director := directerFunc(c, internalCluster, proxyUrl, username, convertedUrl, needConvert, convertedObj)

	errorHandler := func(resp http.ResponseWriter, req *http.Request, err error) {
		if err != nil {
			response.FailReturn(c, errcode.BadRequest(fmt.Errorf("cluster %s url %s proxy fail, %v", cluster, proxyUrl, err)))
			return
		}
	}

	if needConvert {
		// open response filterCondition convert
		_, _, convertedGvr, err := conversion.ParseURL(convertedUrl)
		if err != nil {
			response.FailReturn(c, errcode.BadRequest(err))
			return
		}

		converter, _ := h.converter.GetVersionConvert(cluster)
		converterContext = filter.ConverterContext{
			EnableConvert: true,
			Converter:     converter,
			ConvertedGvr:  convertedGvr,
			RawGvr:        gvr,
		}
	}

	filter := ResponseFilter{
		Condition:        condition,
		ConverterContext: &converterContext,
	}
	needModifyResponse := needModifyResponse(proxyUrl, c)
	// trim auth token here
	c.Request.Header.Del(constants.AuthorizationHeader)
	requestProxy := &httputil.ReverseProxy{Director: director, Transport: transport, ModifyResponse: nil, ErrorHandler: errorHandler}
	if needModifyResponse {
		requestProxy.ModifyResponse = filter.filterResponse
	}
	requestProxy.ServeHTTP(c.Writer, c.Request)
}

func directerFunc(c *gin.Context, internalCluster *multicluster.InternalCluster, proxyUrl, username, convertedUrl string, needConvert bool, convertedObj []byte) func(req *http.Request) {
	return func(req *http.Request) {
		labelSelector := selector.ParseLabelSelector(c.Query("selector"))

		uri, err := url.ParseRequestURI(internalCluster.Config.Host)
		if err != nil {
			response.FailReturn(c, errcode.BadRequest(fmt.Errorf("could not parse host, host: %s , err: %v", internalCluster.Config.Host, err)))
			return
		}
		uri.RawQuery = c.Request.URL.RawQuery
		uri.Path = proxyUrl
		req.URL = uri
		req.Host = internalCluster.Config.Host

		err = requestutil.AddFieldManager(req, username)
		if err != nil {
			clog.Warn("fail to add fieldManager due to %v", err)
		}
		if needConvert {
			// replace request body and url if need
			if convertedObj != nil {
				r := bytes.NewReader(convertedObj)
				body := io.NopCloser(r)
				req.Body = body
				req.ContentLength = int64(r.Len())
			}
			req.URL.Path = convertedUrl
		}

		//In order to improve processing efficiency
		//this method converts requests starting with metadata.labels in the selector into k8s labelSelector requests
		// todo This method can be further optimized and extracted as a function to improve readability
		if len(labelSelector) > 0 {
			convertsLabelSelectorForReq(req, labelSelector)
		}
	}
}

func convertsLabelSelectorForReq(req *http.Request, labelSelector map[string][]string) {
	labelSelectorQueryString := ""
	// Take out the query value in the selector and stitch it into the query field of labelSelector
	// for example: selector=metadata.labels.key=value1|value2|value3
	// then it should be converted to: key+in+(value1,value2,value3)
	for key, value := range labelSelector {
		if len(value) < 1 {
			continue
		}
		labelSelectorQueryString += key
		labelSelectorQueryString += "+in+("
		labelSelectorQueryString += strings.Join(value, ",")
		labelSelectorQueryString += ")"
		labelSelectorQueryString += ","
	}
	if len(labelSelectorQueryString) > 0 {
		labelSelectorQueryString = strings.TrimRight(labelSelectorQueryString, ",")
	}
	labelSelectorQueryString = url.PathEscape(labelSelectorQueryString)
	// Old query parameters may have the following conditions:
	// empty
	// has selector: selector=key=value
	// has selector and labelSelector: selector=key=value&labelSelector=key=value
	// has selector and labelSelector and others: selector=key=value&labelSelector=key=value&fieldSelector=key=value
	// so, use & to split it
	queryArray := strings.Split(req.URL.RawQuery, "&")
	queryString := ""
	labelSelectorSet := false
	for _, v := range queryArray {
		//if it start with labelSelector=, then append converted labelSelector string
		if strings.HasPrefix(v, "labelSelector=") {
			queryString += v + "," + labelSelectorQueryString
			labelSelectorSet = true
			// else if url like: selector=key=value&labelSelector, then use converted labelSelector string replace it
		} else if strings.HasPrefix(v, "labelSelector") {
			queryString += "labelSelector=" + labelSelectorQueryString
			labelSelectorSet = true
			// else no need to do this
		} else {
			queryString += v
		}
		queryString += "&"
	}
	// If the query parameter does not exist labelSelector
	// append converted labelSelector string
	if len(queryString) > 0 && labelSelectorSet == false {
		queryString += "&labelSelector=" + labelSelectorQueryString
	}

	req.URL.RawQuery = queryString
}

// product match/sort/page to other function
func Filter(c *gin.Context, object runtime.Object) (*int, error) {
	condition := parseQueryParams(c)
	total, err := filter.GetEmptyFilter().FilterObjectList(object, condition)
	if err != nil {
		clog.Error("filterCondition userList error, err: %s", err.Error())
		return nil, err
	}
	return &total, nil
}

// parse request params, include selector, sort and page

func parseQueryParams(c *gin.Context) *filter.Condition {
	exact, fuzzy := selector.ParseSelector(c.Query("selector"))
	limit, offset := page.ParsePage(c.Query("pageSize"), c.Query("pageNum"))
	sortName, sortOrder, sortFunc := sort.ParseSort(c.Query("sortName"), c.Query("sortOrder"), c.Query("sortFunc"))
	condition := filter.Condition{
		Exact:     exact,
		Fuzzy:     fuzzy,
		Limit:     limit,
		Offset:    offset,
		SortName:  sortName,
		SortOrder: sortOrder,
		SortFunc:  sortFunc,
	}
	return &condition
}

// if the request has a watch, it should not be filtered
func needModifyResponse(proxyUrl string, c *gin.Context) bool {
	modifyResponse := true
	// According to the rules of k8s, determine whether the request has a watch.
	// When there is a watch, it is a long connection request.
	// At this time, the returned data should not be filtered
	// watch is divided into two cases, one is watch=true in the query parameter, and the other is watch in the path, such as /api/v1/watch/namespaces/default/pods
	// 1. follow=true || watch = true in the query parameter
	follow := c.Query("follow")
	isFollow, _ := strconv.ParseBool(follow)
	if isFollow {
		return false
	}
	watch := c.Query("watch")
	isWatch, _ := strconv.ParseBool(watch)
	if isWatch {
		return false
	}
	// 2. watch in the path
	if len(proxyUrl) > 0 {
		path := proxyUrl
		if strings.HasPrefix(path, "/") {
			path = path[1:]
		}
		split := strings.Split(path, "/")
		if split[0] == "api" {
			if split[2] == "watch" {
				modifyResponse = false
			}
		} else if split[0] == "apis" {
			if split[3] == "watch" {
				modifyResponse = false
			}
		}
	}
	return modifyResponse
}