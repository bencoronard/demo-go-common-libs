package vault

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
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
					token, err := c.authenticate(ctx)
					if err != nil {
						slog.Error("Failed to authenticate with vault server. Re-attempting login.", "error", err)
						select {
						case <-ctx.Done():
							return
						case <-time.After(backoff):
							if backoff < c.cfg.AuthRetryBackoffMaxInterval {
								backoff *= time.Duration(c.cfg.AuthRetryBackoffMult)
							}
							continue
						}
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

func (c *client) autoRenewToken(ctx context.Context, token *vault.Secret) error {
	if !token.Auth.Renewable {
		if token.Auth.LeaseDuration == 0 {
			slog.Info("Token has no TTL and is not renewable. Waiting for context cancellation.")
			<-ctx.Done()
			return nil
		}

		ttl := time.Duration(token.Auth.LeaseDuration) * time.Second
		ttl = ttl * 2 / 3

		slog.Info("Token is not renewable. Re-attempting login before TTL expiry.", "wait", ttl)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(ttl):
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

func (c *client) authenticate(ctx context.Context) (*vault.Secret, error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.ReadTimeout)
	defer cancel()

	var (
		secret *vault.Secret
		err    error
	)

	if c.auth != nil {
		secret, err = c.vc.Auth().Login(ctx, c.auth)
	} else {
		tokenStr, err := c.resolveLocalToken()
		if err != nil {
			return nil, err
		}
		c.vc.SetToken(tokenStr)
		secret, err = c.vc.Auth().Token().LookupSelfWithContext(ctx)
	}

	if err != nil {
		return nil, err
	}
	if secret == nil || secret.Auth == nil {
		return nil, fmt.Errorf("no auth info was returned")
	}

	return secret, nil
}

func (c *client) resolveLocalToken() (string, error) {
	if tokenStr := os.Getenv("VAULT_TOKEN"); tokenStr != "" {
		return tokenStr, nil
	}

	if c.cfg.TokenFilePath != "" {
		data, err := os.ReadFile(c.cfg.TokenFilePath)
		if err != nil {
			return "", err
		}
		if tokenStr := strings.TrimSpace(string(data)); tokenStr != "" {
			return tokenStr, nil
		}
	}

	return "", fmt.Errorf("no local token found")
}
