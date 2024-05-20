package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)



func GetTx(hash string) (*Tx, error) {
	const query = `
	query ($hash: Hash!) {
		tx (hash: $hash) {
			hash
			seen
			raw
			inputs {
				index
				output {
					hash
					index
					amount
					lock {
						address
					}
				}
			}
			outputs {
				index
				amount
				lock {
					address
				}
			}
		}
	}`
	jsonData := map[string]interface{}{
		"query":     query,
		"variables": map[string]interface{}{"hash": hash},
	}
	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json for get tx query; %w", err)
	}
	request, err := http.NewRequest("POST", DefaultUrl, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("error creating new request for get tx; %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: time.Second * 60}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error the graph tx HTTP request failed; %w", err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body; %w", err)
	}
	var dataStruct = struct {
		Data   map[string]Tx `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}{}
	if err := json.Unmarshal(data, &dataStruct); err != nil {
		return nil, fmt.Errorf("error unmarshalling json; %w", err)
	}
	if len(dataStruct.Errors) > 0 {
		return nil, fmt.Errorf("error index client tx response data; %w", fmt.Errorf(dataStruct.Errors[0].Message))
	}
	if tx, ok := dataStruct.Data["tx"]; ok {
		return &tx, nil
	}
	return nil, fmt.Errorf("error getting tx from data")
}
