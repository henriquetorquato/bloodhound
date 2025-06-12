package client

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

type BloodhoundClient struct {
	client      *http.Client
	rateLimiter *rate.Limiter
	config      ClientConfig
}

type ClientConfig struct {
	Rate    int
	Headers map[string]string
	Proxy   string
}

// TODO: Stop when receiving http 429 (Too Many Requests)

func NewClient(config ClientConfig) *BloodhoundClient {
	rateLimiter := rate.NewLimiter(rate.Every(1*time.Second), config.Rate)

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
		rateLimiter: rateLimiter,
		config:      config,
	}
}

func (client *BloodhoundClient) Do(request *http.Request) (*http.Response, error) {
	// Add custom Headers
	for key, value := range client.config.Headers {
		request.Header.Set(key, value)
	}

	// Honour rate limiter
	context := context.Background()
	err := client.rateLimiter.Wait(context)

	if err != nil {
		return nil, err
	}

	response, err := client.client.Do(request)

	if err != nil {
		return nil, err
	}

	return response, nil
}
