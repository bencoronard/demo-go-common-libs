package vault

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/fx"
)

type client struct {
	vc   *vault.Client
	auth vault.AuthMethod
	cfg  Config
}

func (c *client) ReadSecret(ctx context.Context, path string, target any) error {
	readCtx, readCancel := context.WithTimeout(ctx, c.cfg.ReadTimeout)
	defer readCancel()

	secret, err := c.vc.Logical().ReadWithContext(readCtx, path)
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
	ctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				backoff := c.cfg.AuthRetryBackoffInitialInterval
				for {
					if ctx.Err() != nil {
						return
					}
					token, err := c.authenticate(ctx)
					if err != nil {
						slog.Error("Failed to authenticate with vault server", "error", err)
						time.Sleep(backoff)
						backoff = min(
							backoff*time.Duration(c.cfg.AuthRetryBackoffMult),
							c.cfg.AuthRetryBackoffMaxInterval,
						)
						continue
					}
					backoff = c.cfg.AuthRetryBackoffInitialInterval
					if err := c.autoRenewToken(ctx, token); err != nil {
						slog.Error("Failed to renew token", "error", err)
					}
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})

	return nil
}

func (c *client) authenticate(ctx context.Context) (*vault.Secret, error) {
	loginCtx, loginCancel := context.WithTimeout(ctx, c.cfg.ReadTimeout)
	defer loginCancel()

	token, err := c.vc.Auth().Login(loginCtx, c.auth)
	if err != nil {
		return nil, err
	}

	if token == nil || token.Auth == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	return token, nil
}

func (c *client) autoRenewToken(ctx context.Context, token *vault.Secret) error {
	if !token.Auth.Renewable {
		ttl := time.Duration(token.Auth.LeaseDuration) * time.Second
		if ttl == 0 {
			ttl = 1 * time.Minute
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(ttl):
			slog.Info("Token is not configured to be renewable. Re-attempting login.")
			return nil
		}
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
			return nil
		case err := <-watcher.DoneCh():
			if err != nil {
				slog.Info("Failed to renew token. Re-attempting login.", "error", err)
				return nil
			}
			slog.Info("Token can no longer be renewed. Re-attempting login.")
			return nil
		case renewal := <-watcher.RenewCh():
			slog.Info("Token successfully renewed", "data", renewal)
		}
	}
}
