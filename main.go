package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"model"
	"net/http"
	"os"
	"time"

	"github.com/altxtech/warriorplusextractor/model"
)

// WarriorPlusAPIObject interface
type WarriorPlusAPIObject interface {
	model.Sale |
	model.Partner |
	model.Payment |
	model.Affiliate |
	model.Customer |
	model.PartnerList |
	model.PaymentList
}

// Response holds the API response
type Response[T WarriorPlusAPIObject] struct {
	Success    int    `json:"success"`
	Object     string `json:"object"`
	Uri        string `json:"uri"`
	HasMore    bool   `json:"has_more"`
	TotalCount int    `json:"total_count"`
	Data       []T    `json:"data"`
	Errors     Errors `json:"errors"`
}

// Errors holds error information
type Errors struct {
	ErrorType string `json:"error_type"`
	Messages  string `json:"message"`
}

// WarriorPlusClient provides client functionality
type WarriorPlusClient struct {
	ApiKey string
}

// NewWarriorPlusClient creates a new WarriorPlusClient
func NewWarriorPlusClient(apiKey string) *WarriorPlusClient {
	return &WarriorPlusClient{ApiKey: apiKey}
}

// buildHTTPRequest creates an HTTP request
func (client *WarriorPlusClient) buildHTTPRequest(method string, endpoint string) (*http.Request, error) {
	url := "https://warriorplus.com/api/v2" + endpoint
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("apiKey", client.ApiKey)
	req.URL.RawQuery = q.Encode()

	return req, nil
}

// executeHTTPRequest executes an HTTP request
func (client *WarriorPlusClient) executeHTTPRequest(req *http.Request) (*http.Response, error) {
	backoffTime := 100 // Time to back off, in milliseconds
	maxTries := 10

	for i := 0; i < maxTries; i++ {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 200 {
			return resp, nil
		}

		if resp.StatusCode == 429 {
			time.Sleep(time.Duration(backoffTime) * time.Millisecond)
			backoffTime *= 2
			continue
		}

		errorMessage := fmt.Sprintf("HTTP Status Code: %d", resp.StatusCode)
		return resp, errors.New(errorMessage)
	}

	return nil, errors.New("Retry limit exceeded")
}

// parseHTTPResponse parses the HTTP response
func (client *WarriorPlusClient) parseHTTPResponse[T WarriorPlusAPIObject](httpResponse *http.Response) (*Response[T], error) {
	resp := &Response[T]{}

	defer httpResponse.Body.Close()
	responseContent, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(responseContent, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Request makes a request to the API
func (client *WarriorPlusClient) Request[T WarriorPlusAPIObject](method string, endpoint string) (*Response[T], error) {
	req, err := client.buildHTTPRequest(method, endpoint)
	if err != nil {
		return nil, err
	}

	httpResp, err := client.executeHTTPRequest(req)
	if err != nil {
		return nil, err
	}

	resp, err := client.parseHTTPResponse[T](httpResp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func main() {
	log.Println("Creating WarriorPlus Client")
	client := NewWarriorPlusClient(os.Getenv("WARRIORPLUS_API_KEY"))

	// Query sales
	log.Println("Querying sales")
	salesResponse, err := client.Request[Sale]("GET", "/sales")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(salesResponse.Success)
	log.Println(salesResponse.Object)
	log.Println(salesResponse.Uri)
	log.Println(salesResponse.HasMore)
	log.Println(salesResponse.TotalCount)
	log.Println(salesResponse.Data)
}
