package pipeline

import (
	"bloodhound/lib/client"
	"bytes"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/bradhe/stopwatch"
	"golang.org/x/net/html"

	log "github.com/sirupsen/logrus"
)

func RetrieveResource(clientConfig client.ClientConfig, in <-chan Context, out chan<- Context) {
	var wg sync.WaitGroup

	/*
		Each request needs to retrieve a token to be executed,
			and tokens are generated into the channel based on the rate limiting.

		Meaning that every available goroutine will have to wait and respect
			the configured rate limiting
	*/
	tokenChannel := make(chan struct{}, clientConfig.Rate)
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(clientConfig.Rate))
		defer ticker.Stop()

		for {
			<-ticker.C
			tokenChannel <- struct{}{}
		}
	}()

	// TODO: Make this a configuration
	for range 10 {
		wg.Add(1)

		go func() {
			defer wg.Done()
			client := client.NewClient(clientConfig)

			for context := range in {
				// Wait until request is allowed by rate limiter
				<-tokenChannel

				watch := stopwatch.Start()

				log.WithFields(log.Fields{
					"target": context.Url,
				}).Trace("Requesting resource")

				request, err := http.NewRequest("GET", context.Url, nil)

				if err != nil {
					log.WithFields(log.Fields{
						"target": context.Url,
						"err":    err.Error(),
					}).Error("Unable to create HTTP request: Skipping")

					continue
				}

				response, err := client.Do(request)

				if err != nil {
					log.WithFields(log.Fields{
						"target": context.Url,
						"err":    err.Error(),
					}).Error("Unable to process and request URL: Skipping")

					continue
				}

				watch.Stop()

				if response.StatusCode == http.StatusTooManyRequests {
					log.Fatal(`Requests are being limited by target, evaluation received HTTP status 429 Too Many Requests.
						Try running the command again with adjusted request rate settings.`)
				} else if response.StatusCode != http.StatusOK {
					log.WithFields(log.Fields{
						"target":     context.Url,
						"statusCode": response.StatusCode,
						"duration":   watch.Milliseconds(),
					}).Warn("Resource returned non-OK status: Content evaluation will not be available")
				} else {
					log.WithFields(log.Fields{
						"target":   context.Url,
						"duration": watch.Milliseconds(),
					}).Trace("Finished requesting resource")

					// TODO: Check if html.Parse is guaranteed to always read the entire reader
					body, err := io.ReadAll(response.Body)
					response.Body.Close()

					if err != nil {
						log.WithFields(log.Fields{
							"target": context.Url,
							"err":    err,
						}).Error("Unable to read response body")
					}

					bodyReader := bytes.NewReader(body)
					document, _ := html.Parse(bodyReader)
					context.Content = document

					out <- context
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()
}
