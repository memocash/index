package client_test

import (
	"fmt"
	"github.com/memocash/index/db/client"
	"testing"
)

func TestIsMessageNotSetError(t *testing.T) {
	err := fmt.Errorf("test error; %w", client.MessageNotSetError)
	if !client.IsMessageNotSetError(err) {
		t.Error("jerr should be message not set error")
	}
	err2 := fmt.Errorf("wrapped error; %w", client.MessageNotSetError)
	if !client.IsMessageNotSetError(err2) {
		t.Error("fmt.Errorf should be message not set error")
	}
	err3 := fmt.Errorf("other error")
	if client.IsMessageNotSetError(err3) {
		t.Error("other error should not be message not set error")
	}
}
