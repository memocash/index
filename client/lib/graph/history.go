package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"io"
	"net/http"
	"strings"
	"time"
)

type History []AddrTxs

func (h History) GetAllTxs() []Tx {
	var txs []Tx
	for _, addrTxs := range h {
		txs = append(txs, addrTxs.Txs...)
	}
	return txs
}

type AddrTxs struct {
	Address wallet.Addr
	Txs     []Tx
}

func GetHistory(url string, addressUpdates []AddressUpdate) (History, error) {
	jsonValue, err := GetHistoryQuery(addressUpdates)
	if err != nil {
		return nil, jerr.Get("error getting history query", err)
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
		Data   map[string]Address `json:"data"`
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
	var history History
	for _, v := range dataStruct.Data {
		address, err := wallet.GetAddrFromString(v.Address)
		if err != nil {
			return nil, jerr.Get("error getting address from string for history", err)
		}
		history = append(history, AddrTxs{
			Address: *address,
			Txs:     v.Txs,
		})
	}
	return history, nil
}

const txQuery = `{
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
		slp {
			hash
			index
			token_hash
			amount
			genesis {
				hash
				token_type
				decimals
				ticker
				name
				doc_url
			}
		}
		slp_baton {
			hash
			index
			token_hash
			genesis {
				hash
				token_type
				decimals
				ticker
				name
				doc_url
			}
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
					block {
						hash
						timestamp
						height
					}
				}
			}
		}
	}
	blocks {
		block {
			hash
			timestamp
			height
		}
	}
}`

func GetHistoryQuery(addressUpdates []AddressUpdate) ([]byte, error) {
	var variables = make(map[string]interface{})
	var paramsStrings []string
	var subQueries []string
	for i, addressUpdate := range addressUpdates {
		variables[fmt.Sprintf("address%d", i)] = addressUpdate.Address.String()
		variables[fmt.Sprintf("start%d", i)] = addressUpdate.Time.Format(time.RFC3339)
		paramsStrings = append(paramsStrings, fmt.Sprintf("$address%d: String!, $start%d: Date", i, i))
		subQueries = append(subQueries, fmt.Sprintf(`address%d: address(address: $address%d) {
			address
			txs(start: $start%d) %s
		}`, i, i, i, txQuery))
	}
	var query = fmt.Sprintf("query (%s) { %s }", strings.Join(paramsStrings, ", "), strings.Join(subQueries, "\n"))
	jsonData := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}
	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json for get history query; %w", err)
	}
	return jsonValue, nil
}
