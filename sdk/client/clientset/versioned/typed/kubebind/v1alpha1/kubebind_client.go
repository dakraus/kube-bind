/*
Copyright 2024 The Kube Bind Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"net/http"

	rest "k8s.io/client-go/rest"

	v1alpha1 "github.com/kube-bind/kube-bind/sdk/apis/kubebind/v1alpha1"
	"github.com/kube-bind/kube-bind/sdk/client/clientset/versioned/scheme"
)

type KubeBindV1alpha1Interface interface {
	RESTClient() rest.Interface
	APIServiceBindingsGetter
	APIServiceExportsGetter
	APIServiceExportRequestsGetter
	APIServiceNamespacesGetter
	ClusterBindingsGetter
}

// KubeBindV1alpha1Client is used to interact with features provided by the kube-bind.io group.
type KubeBindV1alpha1Client struct {
	restClient rest.Interface
}

func (c *KubeBindV1alpha1Client) APIServiceBindings() APIServiceBindingInterface {
	return newAPIServiceBindings(c)
}

func (c *KubeBindV1alpha1Client) APIServiceExports(namespace string) APIServiceExportInterface {
	return newAPIServiceExports(c, namespace)
}

func (c *KubeBindV1alpha1Client) APIServiceExportRequests(namespace string) APIServiceExportRequestInterface {
	return newAPIServiceExportRequests(c, namespace)
}

func (c *KubeBindV1alpha1Client) APIServiceNamespaces(namespace string) APIServiceNamespaceInterface {
	return newAPIServiceNamespaces(c, namespace)
}

func (c *KubeBindV1alpha1Client) ClusterBindings(namespace string) ClusterBindingInterface {
	return newClusterBindings(c, namespace)
}

// NewForConfig creates a new KubeBindV1alpha1Client for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*KubeBindV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	httpClient, err := rest.HTTPClientFor(&config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&config, httpClient)
}

// NewForConfigAndClient creates a new KubeBindV1alpha1Client for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*KubeBindV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &KubeBindV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new KubeBindV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *KubeBindV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new KubeBindV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *KubeBindV1alpha1Client {
	return &KubeBindV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *KubeBindV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}