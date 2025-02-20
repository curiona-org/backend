package redis

import (
	"context"

	baseredis "github.com/redis/go-redis/v9"
)

type Client struct {
	*baseredis.Client
}

func New(ctx context.Context, cfg *Config) (*Client, error) {
	rdb := baseredis.NewClient(optionsFromConfig(cfg))

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{rdb}, nil
}

func optionsFromConfig(cfg *Config) *baseredis.Options {
	opts := &baseredis.Options{
		Network:               cfg.Network,
		Addr:                  cfg.Addr,
		DB:                    cfg.DB,
		Protocol:              3,
		Username:              cfg.Username,
		Password:              cfg.Password,
		ContextTimeoutEnabled: true,
	}

	return opts
}

func (c *Client) NewScript(src string) *baseredis.Script {
	return baseredis.NewScript(src)
}
