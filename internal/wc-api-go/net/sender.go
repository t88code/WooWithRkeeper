package net // import "WooWithRkeeper/internal/wc-api-go/net"

import (
	"bytes"
	"encoding/json"
	"net/http"

	"WooWithRkeeper/internal/wc-api-go/request"
)

// Sender provides HTTP Requests
type Sender struct {
	requestEnricher RequestEnricher
	urlBuilder      URLBuilder
	httpClient      Client
	requestCreator  RequestCreator
}

// Send method sends requests to WooCommerce API
func (s *Sender) Send(req request.Request) (resp *http.Response, err error) {
	request := s.prepareRequest(req)
	return s.httpClient.Do(request)
}

func (s *Sender) prepareRequest(req request.Request) *http.Request {
	URL := s.urlBuilder.GetURL(req)

	reqBody, err := json.Marshal(req.Body)
	if err != nil || (req.Method != "POST" && req.Method != "PUT") {
		reqBody = nil
	}

	request, _ := s.requestCreator.NewRequest(req.Method, URL, bytes.NewBuffer(reqBody))
	s.requestEnricher.EnrichRequest(request, URL)
	request.Header.Set("Content-Type", "application/json")
	return request
}

// SetRequestEnricher ...
func (s *Sender) SetRequestEnricher(a RequestEnricher) {
	s.requestEnricher = a
}

// SetURLBuilder ...
func (s *Sender) SetURLBuilder(urlBuilder URLBuilder) {
	s.urlBuilder = urlBuilder
}

// SetHTTPClient ...
func (s *Sender) SetHTTPClient(c Client) {
	s.httpClient = c
}

// SetRequestCreator ...
func (s *Sender) SetRequestCreator(rc RequestCreator) {
	s.requestCreator = rc
}
