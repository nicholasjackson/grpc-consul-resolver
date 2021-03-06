package echo

import (
	"fmt"

	context "golang.org/x/net/context"
)

type EchoServiceServerImpl struct {
	ID string
}

func (e *EchoServiceServerImpl) Echo(ctx context.Context, in *Message) (*Message, error) {
	return &Message{Data: fmt.Sprintf("Server: %s, Message: %s", e.ID, in.Data)}, nil
}
