package network_client

import (
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/ref/network/gen/network_pb"
)

type OutputMessenger struct {
}

func (m *OutputMessenger) Output(message string) error {
	connection, err := NewConnection()
	if err != nil {
		return jerr.Get("error connecting to network", err)
	}
	defer connection.Close()
	if _, err := connection.Client.OutputMessage(connection.GetDefaultContext(), &network_pb.StringMessage{
		Message: message,
	}); err != nil {
		return jerr.Get("could not send output message to network", err)
	}
	return nil
}

func NewOutputMessenger() *OutputMessenger {
	return &OutputMessenger{}
}
