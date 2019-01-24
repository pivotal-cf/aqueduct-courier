package cf

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type OAuthClient struct {
	oauthConfig   *oauth2.Config
	oauthConfigCC *clientcredentials.Config
	context       context.Context
	target        string
	timeout       time.Duration
}

type client interface {
	Do(request *http.Request) (*http.Response, error)
}

func NewOAuthClient(target, clientID, clientSecret string, requestTimeout time.Duration, client client) OAuthClient {
	confCC := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	return funcName(confCC, client, target, requestTimeout)
}

func funcName(confCC *clientcredentials.Config, client client, target string, requestTimeout time.Duration) OAuthClient {
	return OAuthClient{
		oauthConfig:   &oauth2.Config{},
		oauthConfigCC: confCC,
		context:       context.WithValue(context.TODO(), oauth2.HTTPClient, client),
		target:        target,
		timeout:       requestTimeout,
	}
}

func (oc OAuthClient) Do(request *http.Request) (*http.Response, error) {
	var client *http.Client

	targetURL, err := url.Parse(oc.target)
	if err != nil {
		return nil, fmt.Errorf("could not parse target url: %s", err)
	}

	targetURL.Path = "/oauth/token"
	oc.oauthConfigCC.TokenURL = targetURL.String()
	oc.oauthConfig.Endpoint.TokenURL = targetURL.String()

	client = oc.oauthConfigCC.Client(oc.context)
	client.Timeout = oc.timeout

	resp, err := client.Do(request)

	if err != nil {
		return nil, fmt.Errorf("error performing request %s", err)
	}
	return resp, err
}
