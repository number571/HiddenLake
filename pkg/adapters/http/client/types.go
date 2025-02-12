package client

import (
	"context"

	"github.com/number571/go-peer/pkg/message/layer1"
)

type IClient interface {
	GetIndex(context.Context) (string, error)

	GetOnlines(context.Context) ([]string, error)
	DelOnline(context.Context, string) error

	GetConnections(context.Context) ([]string, error)
	AddConnection(context.Context, string) error
	DelConnection(context.Context, string) error

	ProduceMessage(context.Context, layer1.IMessage) error
}

type IRequester interface {
	GetIndex(context.Context) (string, error)

	GetOnlines(context.Context) ([]string, error)
	DelOnline(context.Context, string) error

	GetConnections(context.Context) ([]string, error)
	AddConnection(context.Context, string) error
	DelConnection(context.Context, string) error

	ProduceMessage(context.Context, layer1.IMessage) error
}
