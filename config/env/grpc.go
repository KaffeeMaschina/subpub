package env

import (
	"errors"
	"github.com/KaffeeMaschina/subpub/internal/config"
	"net"
	"os"
)

// grpcConfig should have methode Address() string to implement config.GRPCConfig
var _ config.GRPCConfig = (*grpcConfig)(nil)

type grpcConfig struct {
	host string
	port string
}

func NewGRPCConfig() (*grpcConfig, error) {
	host := os.Getenv("GRPC_HOST")
	if len(host) == 0 {
		return nil, errors.New("grpc host not found")
	}
	port := os.Getenv("GRPC_PORT")
	if len(port) == 0 {
		return nil, errors.New("grpc port not found")
	}
	return &grpcConfig{
		host: host,
		port: port,
	}, nil
}

func (c *grpcConfig) Address() string {
	return net.JoinHostPort(c.host, c.port)
}
