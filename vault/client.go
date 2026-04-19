package vault

import (
	"context"
	"encoding/json"
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

type watcherParams struct {
	fx.In
	Lc  fx.Lifecycle
	Cfg Config
}

func (c *client) WatchTokenLifecycle(p watcherParams) error {
	ctx, cancel := context.WithCancel(context.Background())

	p.Lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				backoff := p.Cfg.AuthRetryBackoffInitialInterval
				for {
					authCtx, authCancel := context.WithTimeout(ctx, p.Cfg.AuthTimeout)
					token, err := c.authenticate(authCtx)
					authCancel()
					if err != nil {
						slog.Error("Failed to authenticate with vault server. Re-attempting login.", "error", err)
						select {
						case <-ctx.Done():
							return
						case <-time.After(backoff):
							if backoff < p.Cfg.AuthRetryBackoffMaxInterval {
								backoff *= time.Duration(p.Cfg.AuthRetryBackoffMultiplier)
							}
							continue
						}
					}
					backoff = p.Cfg.AuthRetryBackoffInitialInterval
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

func (c *client) autoRenewToken(ctx context.Context, s *vault.Secret) error {
	if !isRenewable(s) {
		ttl := getTTL(s)

		if ttl <= 0 {
			slog.Info("Token has no expiration. Waiting for context cancellation.")
			<-ctx.Done()
			return nil
		}

		wait := ttl * 2 / 3
		slog.Info("Token is static. Re-logging after grace period.", "wait", wait)

		select {
		case <-ctx.Done():
		case <-time.After(wait):
		}

		return nil
	}

	watcher, err := c.vc.NewLifetimeWatcher(&vault.LifetimeWatcherInput{Secret: s})
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
			slog.Info("Watcher finished. Re-attempting login.", "error", err)
			return nil
		case renewal := <-watcher.RenewCh():
			slog.Info("Token successfully renewed", "data", renewal)
		}
	}
}

func (c *client) authenticate(ctx context.Context) (*vault.Secret, error) {
	var (
		secret *vault.Secret
		err    error
	)

	if c.auth != nil {
		secret, err = c.vc.Auth().Login(ctx, c.auth)
	} else {
		if err := c.resolveLocalToken(); err != nil {
			return nil, err
		}
		secret, err = c.vc.Auth().Token().LookupSelfWithContext(ctx)
	}

	if err != nil {
		return nil, err
	}

	if secret == nil {
		return nil, fmt.Errorf("vault returned an empty response")
	}

	isTokenLookup := secret.Data != nil && secret.Data["id"] != nil

	if secret.Auth == nil && !isTokenLookup {
		return nil, fmt.Errorf("no valid authentication or token metadata found")
	}

	return secret, nil
}

func (c *client) resolveLocalToken() error {
	if tokenStr := strings.TrimSpace(os.Getenv("VAULT_TOKEN")); tokenStr != "" {
		c.vc.SetToken(tokenStr)
		return nil
	}

	path := strings.TrimSpace(os.Getenv("VAULT_TOKEN_FILE"))
	if path == "" {
		return fmt.Errorf("no vault token found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	tokenStr := strings.TrimSpace(string(data))
	if tokenStr == "" {
		return fmt.Errorf("token file at %s was empty", path)
	}

	c.vc.SetToken(tokenStr)

	return nil
}

func isRenewable(s *vault.Secret) bool {
	if s == nil {
		return false
	}
	if s.Renewable {
		return true
	}
	if s.Auth != nil && s.Auth.Renewable {
		return true
	}
	return false
}

func getTTL(s *vault.Secret) time.Duration {
	if s == nil {
		return 0
	}
	if s.Auth != nil && s.Auth.LeaseDuration > 0 {
		return time.Duration(s.Auth.LeaseDuration) * time.Second
	}
	if s.LeaseDuration > 0 {
		return time.Duration(s.LeaseDuration) * time.Second
	}
	if ttlVal, ok := s.Data["ttl"].(int); ok {
		return time.Duration(ttlVal) * time.Second
	}
	if ttlVal, ok := s.Data["ttl"].(json.Number); ok {
		t, _ := ttlVal.Int64()
		return time.Duration(t) * time.Second
	}
	return 0
}
