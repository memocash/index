package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"io/ioutil"
	"net/http"
	"time"
)

func GetHistory(address *wallet.Addr, startHeight int64) ([]Tx, error) {
	jsonData := map[string]string{
		"query": HistoryQuery,
		"variables": fmt.Sprintf(`{
			"address": "%s",
			"height": %d
			}`, address.String(), startHeight),
	}
	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest("POST", "http://localhost:26770/graphql", bytes.NewBuffer(jsonValue))
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
	}{}
	if err := json.Unmarshal(data, &dataStruct); err != nil {
		return nil, jerr.Get("error unmarshalling json", err)
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

const HistoryQuery = `{
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
