package client

import (
	"net/http"
	"net/url"
)

type BloodhoundClient struct {
	client *http.Client
	config ClientConfig
}

type ClientConfig struct {
	Rate    int
	Headers map[string]string
	Proxy   string
}

func NewClient(config ClientConfig) *BloodhoundClient {
	return &BloodhoundClient{
		client: &http.Client{
			Transport: &http.Transport{
				Proxy: func(r *http.Request) (*url.URL, error) {
					if config.Proxy != "" {
						proxy, err := url.Parse(config.Proxy)

						if err != nil {
							return nil, err
						}

						return proxy, nil
					} else {
						return nil, nil
					}
				},
			},
		},
		config: config,
	}
}

func (client *BloodhoundClient) Do(request *http.Request) (*http.Response, error) {
	// Add custom Headers
	for key, value := range client.config.Headers {
		request.Header.Set(key, value)
	}

	response, err := client.client.Do(request)

	if err != nil {
		return nil, err
	}

	return response, nil
}
