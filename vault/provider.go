package vault

import (
	"context"
	"time"

	"go.uber.org/fx"

	vault "github.com/hashicorp/vault/api"
)

type Client interface {
	ReadSecret(ctx context.Context, path string, target any) error
	WatchTokenLifecycle(p watcherParams) error
}

type ClientConfig struct {
	TokenFilePath string
}

type WatcherConfig struct {
	AuthTimeout                     time.Duration
	AuthRetryBackoffInitialInterval time.Duration
	AuthRetryBackoffMultiplier      int
	AuthRetryBackoffMaxInterval     time.Duration
}

type clientParams struct {
	fx.In
	Lc   fx.Lifecycle
	Auth vault.AuthMethod `optional:"true"`
	Cfg  ClientConfig
}

func NewClient(p clientParams) (Client, error) {
	cfg := vault.DefaultConfig()
	if cfg.Error != nil {
		return nil, cfg.Error
	}

	vc, err := vault.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	c := client{vc: vc, auth: p.Auth, cfg: p.Cfg}

	p.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if _, err := c.authenticate(ctx); err != nil {
				return err
			}
			return nil
		},
	})

	return &c, nil
}
