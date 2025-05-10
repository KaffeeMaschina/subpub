package pubsubclient

import (
	"context"
	"fmt"
	"log"

	desc "github.com/KaffeeMaschina/subpub/pkg/bus_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn   *grpc.ClientConn
	client desc.PubSubClient
}

func NewClient(address string) *Client {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to connect to server")
	}
	client := desc.NewPubSubClient(conn)
	return &Client{
		conn:   conn,
		client: client,
	}
}

func (c *Client) Publish(ctx context.Context, key string, data string) error {
	_, err := c.client.Publish(ctx, &desc.PublishRequest{
		Key:  key,
		Data: data,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return fmt.Errorf("unknown error: %v", err)
		}
		switch st.Code() {
		case codes.InvalidArgument:
			return fmt.Errorf("invalid argument: %s", st.Message())
		case codes.Internal:
			return fmt.Errorf("internal server error: %s", st.Message())
		case codes.Unavailable:
			return fmt.Errorf("service unavailable: %s", st.Message())
		default:
			return fmt.Errorf("unexpected error: %s", st.Message())
		}
	}
	return nil
}
func (c *Client) Subscribe(ctx context.Context, key string, handler func(data string)) error {
	stream, err := c.client.Subscribe(ctx, &desc.SubscribeRequest{Key: key})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return fmt.Errorf("unknown error: %v", err)
		}
		switch st.Code() {
		case codes.InvalidArgument:
			return fmt.Errorf("invalid argument: %s", st.Message())
		case codes.Internal:
			return fmt.Errorf("internal server error: %s", st.Message())
		case codes.Unavailable:
			return fmt.Errorf("service unavailable: %s", st.Message())
		default:
			return fmt.Errorf("unexpected error: %s", st.Message())
		}
	}

	go func() {
		for {
			event, err := stream.Recv()
			if err != nil {
				st, ok := status.FromError(err)
				if !ok {
					log.Printf("unknown stream error: %v", err)
					return
				}

				switch st.Code() {
				case codes.Canceled:
					log.Printf("stream canceled: %s", st.Message())
				case codes.Unavailable:
					log.Printf("service unavailable: %s", st.Message())
				default:
					log.Printf("stream error: %s", st.Message())
				}
				return
			}

			handler(event.Data)
		}
	}()
	return nil
}
func (c *Client) Close() error {
	return c.conn.Close()
}
