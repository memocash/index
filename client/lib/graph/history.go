package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"io"
	"net/http"
	"time"
)

type History []Tx

func GetHistory(url string, address wallet.Addr, lastUpdate time.Time) ([]Tx, error) {
	jsonData := map[string]interface{}{
		"query": HistoryQuery,
		"variables": map[string]interface{}{
			"address": address.String(),
			"start":   lastUpdate.Format(time.RFC3339),
		},
	}
	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json for get history; %w", err)
	}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("error creating new request for get history; %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error the HTTP request failed; %w", err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body; %w", err)
	}
	var dataStruct = struct {
		Data struct {
			Address Address `json:"address"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}{}
	if err := json.Unmarshal(data, &dataStruct); err != nil {
		return nil, fmt.Errorf("error unmarshalling json; %w", err)
	}
	if len(dataStruct.Errors) > 0 {
		return nil, fmt.Errorf("error index client history response data; %w", fmt.Errorf(dataStruct.Errors[0].Message))
	}
	return dataStruct.Data.Address.Txs, nil
}

const QueryTx = `{
	hash
	seen
	raw
	inputs {
		index
		prev_hash
		prev_index
	}
	outputs {
		index
		amount
		lock {
			address
		}
		spends {
			tx {
				hash
				seen
				raw
				inputs {
					index
					prev_hash
					prev_index
				}
				outputs {
					index
					amount
					lock {
						address
					}
				}
				blocks {
					hash
					timestamp
					height
				}
			}
		}
	}
	blocks {
		hash
		timestamp
		height
	}
}`

const HistoryQuery = `query ($address: String!, $start: Date) {
	address (address: $address) {
		txs(start: $start) ` + QueryTx + `
	}
}`
