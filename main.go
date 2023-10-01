package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	"os"
	"io"
)

/*
	Overview.
	I want this to be the prototype service for everything else

	Client -> Abstract Away http

	Extractor -> Uses the client to extract stuff

	API -> Receives extraction requests
*/

// Types
type WarriorPlusClient struct {
	ApiKey string
}

func NewWarriorPlusClient( apiKey string ) *WarriorPlusClient {
	return &WarriorPlusClient{ ApiKey: apiKey }
}

func (client *WarriorPlusClient) buildHTTPRequest( method string, endpoint string ) ( *http.Request, error ){
	url := "https://warriorplus.com/api/v2" + endpoint
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// Add the apiKey parameter
	q := req.URL.Query()
	q.Add("apiKey", client.ApiKey)
	req.URL.RawQuery = q.Encode()

	return req, nil
}

func (client *WarriorPlusClient) executeHTTPRequest(req *http.Request) (*http.Response, error) {
	/*
		Execute the request with exponential backoffTime
	*/

	backoffTime := 100 // Time to backkoff, in milliseconds 
	maxTries := 10


	for i := 0; i < maxTries; i++ {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		// Return early in case of success 
		if resp.StatusCode == 200 {
			return resp, nil
		}

		// Backoff in case of rate limit exceeded
		if resp.StatusCode == 429 {
			time.Sleep(time.Duration(backoffTime) * time.Millisecond)
			backoffTime *= 2
			continue
		}

		// Return errors for any other status StatusCode
		errorMessage := fmt.Sprintf("HTTP Status Code: %d", resp.StatusCode)
		return resp, errors.New(errorMessage)
	}
	
	return nil, errors.New("Retry limit exceeded") 
}

func (client *WarriorPlusClient ) parseHTTPResponse (httpResponse *http.Response) (*Response, error){

	resp := &Response{}

	// Parse the Response
	defer httpResponse.Body.Close()
	responseContent, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(responseContent, resp)
	return resp, nil
}

func (client *WarriorPlusClient) Request(method string, endpoint string) ( *Response, error ) {
	
	// Build request for the endpoint
	req, err := client.buildHTTPRequest(method, endpoint)
	log.Println(req.URL)
	if err != nil {
		return nil, err
	}
	// Execute the http request
	httpResp, err := client.executeHTTPRequest(req)
	if err != nil {
		return nil, err
	}
	// Parse the response
	resp, err := client.parseHTTPResponse(httpResp)
	if err != nil {
		return nil, err
	}
	
	return resp, nil
}
	
type Errors struct {
	ErrorType string `json:"error_type"`
	Messages string `json:"message"`
}

type Response struct {
	Success int `json:"success"`
	Object string `json:"object"`
	Uri string `json:"uri"`
	HasMore bool `json:"has_more"`
	TotalCount int `json:"total_count"`
	Data []interface{} `json:"data"`
	Errors Errors `json:"errors"`
}


func main() {
	log.Println("Creating WarriorPlus Client")
	client := NewWarriorPlusClient(os.Getenv("WARRIORPLUS_API_KEY"))

	// Query sales
	log.Println("Querying sales")
	salesResponse, err := client.Request("GET", "/sales")
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
