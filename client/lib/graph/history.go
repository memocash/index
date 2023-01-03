package graph

import (
	"bytes"
	"encoding/json"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"io/ioutil"
	"net/http"
	"time"
)

func GetHistory(address *wallet.Addr, startHeight int64) ([]Tx, error) {
	jsonData := map[string]interface{}{
		"query": HistoryQuery,
		"variables": map[string]interface{}{
			"address": address.String(),
			"height":  startHeight,
		},
	}
	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return nil, jerr.Get("error marshaling json for get history", err)
	}
	request, err := http.NewRequest("POST", "http://localhost:26770/graphql", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, jerr.Get("error creating new request for get history", err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	if err != nil {
		return nil, jerr.Get("error the HTTP request failed", err)
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, jerr.Get("error reading response body", err)
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
		return nil, jerr.Get("error unmarshalling json", err)
	}
	if len(dataStruct.Errors) > 0 {
		return nil, jerr.Get("error response data", jerr.New(dataStruct.Errors[0].Message))
	}
	var txs []Tx
OutputLoop:
	for _, output := range dataStruct.Data.Address.Outputs {
		for _, tx := range txs {
			if tx.Hash == output.Tx.Hash {
				continue OutputLoop
			}
		}
		txs = append(txs, output.Tx)
	}
	return txs, nil
}

const HistoryQuery = `query ($address: String!, $height: Int) {
	address (address: $address) {
		outputs(height: $height) {
			hash
			index
			amount
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
			}
		}
	}
}`
