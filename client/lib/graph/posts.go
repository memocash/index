package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func GetPosts(start time.Time) ([]*Post, error) {
	const query = `
	query ($start: Date) {
		posts_newest (start: $start) {
			address
			tx_hash
			text
			tx {
				seen
			}
			room {
			    name
			}
		}
	}`
	jsonData := map[string]interface{}{
		"query":     query,
		"variables": map[string]interface{}{"start": start.Format(time.RFC3339)},
	}
	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json for get posts query; %w", err)
	}
	request, err := http.NewRequest("POST", DefaultUrl, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("error creating new request for get posts; %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: time.Second * 60}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error the graph posts HTTP request failed; %w", err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body posts; %w", err)
	}
	var dataStruct = struct {
		Data   map[string][]*Post `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}{}
	if err := json.Unmarshal(data, &dataStruct); err != nil {
		return nil, fmt.Errorf("error unmarshalling posts json; %w", err)
	}
	if len(dataStruct.Errors) > 0 {
		return nil, fmt.Errorf("error index client posts response data; %w", fmt.Errorf("%s", dataStruct.Errors[0].Message))
	}
	if posts, ok := dataStruct.Data["posts_newest"]; ok {
		return posts, nil
	}
	return nil, fmt.Errorf("error getting posts")
}
