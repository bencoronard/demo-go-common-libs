package vault

import (
	"context"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/fx"
)

type client struct {
	vc   *vault.Client
	auth vault.AuthMethod
}

func (c *client) ReadSecret(ctx context.Context, path string, target any) error {
	secret, err := c.vc.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return err
	}
	if secret == nil || secret.Data == nil {
		return fmt.Errorf("%w: secret not found at path: %s", ErrSecretNotFound, path)
	}

	data := secret.Data
	if nested, ok := secret.Data["data"].(map[string]any); ok {
		data = nested
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           target,
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
	})
	if err != nil {
		return err
	}

	return decoder.Decode(data)
}

func (c *client) WatchTokenLifecycle(lc fx.Lifecycle) error {
	renewCtx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go c.runTokenRenewalLoop(renewCtx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			cancel()
			return nil
		},
	})

	return nil
}

func (c *client) login(ctx context.Context) (*vault.Secret, error) {
	authInfo, err := c.vc.Auth().Login(ctx, c.auth)
	if err != nil {
		return nil, err
	}
	if authInfo == nil {
		return nil, fmt.Errorf("%w: no auth info returned", ErrAuthenticationFail)
	}
	return authInfo, nil
}

func (c *client) runTokenRenewalLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := c.manageTokenLifecycle(ctx); err != nil {
				return
			}
			if _, err := c.login(ctx); err != nil {
				continue
			}
		}
	}
}

func (c *client) manageTokenLifecycle(ctx context.Context) error {
	token, err := c.vc.Auth().Token().LookupSelfWithContext(ctx)
	if err != nil {
		return err
	}

	if !token.Auth.Renewable {
		return nil
	}

	watcher, err := c.vc.NewLifetimeWatcher(&vault.LifetimeWatcherInput{Secret: token})
	if err != nil {
		return err
	}

	go watcher.Start()
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-watcher.DoneCh():
			if err != nil {
				return nil
			}
			return nil
		case <-watcher.RenewCh():
		}
	}
}
