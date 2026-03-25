package keycloak

import (
	"context"
	//"github.com/keycloak-broker/pkg/catalog"
	//"github.com/keycloak-broker/pkg/logger"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) CreateClient(ctx context.Context, instanceId, serviceId, planId string) (string, error) {
	return instanceId, nil
}

func (c *Client) GetClient(ctx context.Context, instanceId string) (string, error) {
	return instanceId, nil
}

func (c *Client) DeleteClient(ctx context.Context, instanceId string) error {
	return nil
}
