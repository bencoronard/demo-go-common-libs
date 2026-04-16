package vault

import (
	"context"
	"time"

	"go.uber.org/fx"

	vault "github.com/hashicorp/vault/api"
)

type Client interface {
	ReadSecret(ctx context.Context, path string, target any) error
	WatchTokenLifecycle(lc fx.Lifecycle) error
}

type Config struct {
	ReadTimeout                     time.Duration
	AuthTimeout                     time.Duration
	AuthRetryBackoffInitialInterval time.Duration
	AuthRetryBackoffMult            int
	AuthRetryBackoffMaxInterval     time.Duration
	AuthDefaultTtl                  time.Duration
}

type Params struct {
	fx.In
	Lc   fx.Lifecycle
	Auth vault.AuthMethod `optional:"true"`
	Cfg  Config
}

func NewClient(p Params) (Client, error) {
	cfg := vault.DefaultConfig()
	if cfg.Error != nil {
		return nil, cfg.Error
	}

	vc, err := vault.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	c := client{vc: vc, auth: p.Auth, cfg: p.Cfg}

	if p.Auth == nil {
		return &c, nil
	}

	authInfo, err := c.vc.Auth().Login(context.Background(), p.Auth)
	if err != nil {
		return nil, err
	}

	watcher, err := c.vc.NewLifetimeWatcher(&vault.LifetimeWatcherInput{Secret: authInfo})
	if err != nil {
		return nil, err
	}

	go watcher.Start()
	defer watcher.Stop()

	return &c, nil
}
