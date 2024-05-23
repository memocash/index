package network_client

import (
	"fmt"
	"github.com/memocash/index/ref/network/gen/network_pb"
)

type OutputMessenger struct {
}

func (m *OutputMessenger) Output(message string) error {
	connection, err := NewConnection()
	if err != nil {
		return fmt.Errorf("error connecting to network; %w", err)
	}
	defer connection.Close()
	if _, err := connection.Client.OutputMessage(connection.GetDefaultContext(), &network_pb.StringMessage{
		Message: message,
	}); err != nil {
		return fmt.Errorf("could not send output message to network; %w", err)
	}
	return nil
}

func NewOutputMessenger() *OutputMessenger {
	return &OutputMessenger{}
}
