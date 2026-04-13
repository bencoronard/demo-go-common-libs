package vault

import (
	"context"

	"go.uber.org/fx"

	vault "github.com/hashicorp/vault/api"
)

type Client interface {
	ReadSecret(ctx context.Context, path string, target any) error
	WatchTokenLifecycle(lc fx.Lifecycle) error
}

type Params struct {
	fx.In
	Lc   fx.Lifecycle
	Auth vault.AuthMethod `optional:"true"`
}

func NewTokenClient(p Params) (Client, error) {
	cfg := vault.DefaultConfig()
	if cfg.Error != nil {
		return nil, cfg.Error
	}

	vc, err := vault.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	c := client{vc: vc}

	p.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if _, err := vc.Auth().Token().LookupSelfWithContext(ctx); err != nil {
				return err
			}
			return nil
		},
	})

	return &c, nil
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

	c := client{vc: vc, auth: p.Auth}

	p.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if _, err := c.login(ctx); err != nil {
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return c.vc.Auth().Token().RevokeSelfWithContext(ctx, "")
		},
	})

	return &c, nil
}
