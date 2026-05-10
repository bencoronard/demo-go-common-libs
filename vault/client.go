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
	client *vault.Client
	auth   vault.AuthMethod
}

func (c *client) ReadSecret(ctx context.Context, path string, target any) error {
	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read secret at path %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return fmt.Errorf("secret not found at path: %s", path)
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
		return fmt.Errorf("failed to decode secret: %w", err)
	}

	return decoder.Decode(data)
}

type watcherParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Config    Config
}

func (c *client) WatchTokenLifecycle(p watcherParams) error {
	ctx, cancel := context.WithCancel(context.Background())

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				backoff := p.Config.AuthRetryBackoffInitialInterval
				for {
					authCtx, authCancel := context.WithTimeout(ctx, p.Config.AuthTimeout)
					token, err := c.authenticate(authCtx)
					authCancel()
					if err != nil {
						slog.Error("authentication failed", "error", err)
						select {
						case <-ctx.Done():
							return
						case <-time.After(backoff):
							if backoff < p.Config.AuthRetryBackoffMaxInterval {
								backoff *= time.Duration(p.Config.AuthRetryBackoffMultiplier)
							}
							continue
						}
					}
					backoff = p.Config.AuthRetryBackoffInitialInterval
					if err := c.autoRenewToken(ctx, token); err != nil {
						slog.Error("renewal failed", "error", err)
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
			<-ctx.Done()
			return nil
		}
		wait := ttl * 2 / 3
		select {
		case <-ctx.Done():
		case <-time.After(wait):
		}
		return nil
	}

	watcher, err := c.client.NewLifetimeWatcher(&vault.LifetimeWatcherInput{Secret: s})
	if err != nil {
		return fmt.Errorf("failed to create a lifetime watcher: %w", err)
	}

	go watcher.Start()
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-watcher.DoneCh():
			return nil
		case <-watcher.RenewCh():
		}
	}
}

func (c *client) authenticate(ctx context.Context) (*vault.Secret, error) {
	var (
		secret *vault.Secret
		err    error
	)

	if c.auth != nil {
		secret, err = c.client.Auth().Login(ctx, c.auth)
	} else {
		if err := c.resolveLocalToken(); err != nil {
			return nil, fmt.Errorf("failed to resolve token: %w", err)
		}
		secret, err = c.client.Auth().Token().LookupSelfWithContext(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	if secret == nil {
		return nil, fmt.Errorf("received empty response")
	}

	isTokenLookup := secret.Data != nil && secret.Data["id"] != nil

	if secret.Auth == nil && !isTokenLookup {
		return nil, fmt.Errorf("no valid authentication or token metadata found")
	}

	return secret, nil
}

func (c *client) resolveLocalToken() error {
	if tokenStr := strings.TrimSpace(os.Getenv("VAULT_TOKEN")); tokenStr != "" {
		c.client.SetToken(tokenStr)
		return nil
	}

	path := strings.TrimSpace(os.Getenv("VAULT_TOKEN_FILE"))
	if path == "" {
		return fmt.Errorf("no vault token found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read token file at path %s: %w", path, err)
	}

	tokenStr := strings.TrimSpace(string(data))
	if tokenStr == "" {
		return fmt.Errorf("token file at %s was empty", path)
	}

	c.client.SetToken(tokenStr)

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
