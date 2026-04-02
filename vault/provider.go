package vault

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/fx"

	vault "github.com/hashicorp/vault/api"
	authAppRole "github.com/hashicorp/vault/api/auth/approle"
	authK8s "github.com/hashicorp/vault/api/auth/kubernetes"
	authUsrPsw "github.com/hashicorp/vault/api/auth/userpass"
)

type Client interface {
	ReadSecret(ctx context.Context, path string, target any) error
	WatchTokenLifecycle(lc fx.Lifecycle) error
}

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

type K8sClientConfig struct {
	VaultAddr      string
	RoleName       string
	TokenMountPath string
}

type K8sClientParams struct {
	fx.In
	Lc  fx.Lifecycle
	Cfg *K8sClientConfig
}

func NewK8sClient(p K8sClientParams) (Client, error) {
	auth, err := authK8s.NewKubernetesAuth(p.Cfg.RoleName, authK8s.WithServiceAccountTokenPath(p.Cfg.TokenMountPath))
	if err != nil {
		return nil, err
	}
	return newClient(p.Lc, p.Cfg.VaultAddr, auth)
}

type AppRoleClientConfig struct {
	VaultAddr string
	RoleID    string
	SecretID  string
}

type AppRoleClientParams struct {
	fx.In
	Lc  fx.Lifecycle
	Cfg *AppRoleClientConfig
}

func NewAppRoleClient(p AppRoleClientParams) (Client, error) {
	auth, err := authAppRole.NewAppRoleAuth(p.Cfg.RoleID, &authAppRole.SecretID{FromString: p.Cfg.SecretID})
	if err != nil {
		return nil, err
	}
	return newClient(p.Lc, p.Cfg.VaultAddr, auth)
}

type UserPassClientConfig struct {
	VaultAddr string
	Username  string
	Password  string
}

type UserPassClientParams struct {
	fx.In
	Lc  fx.Lifecycle
	Cfg *UserPassClientConfig
}

func NewUserPassClient(p UserPassClientParams) (Client, error) {
	auth, err := authUsrPsw.NewUserpassAuth(p.Cfg.Username, &authUsrPsw.Password{FromString: p.Cfg.Password})
	if err != nil {
		return nil, err
	}
	return newClient(p.Lc, p.Cfg.VaultAddr, auth)
}

type TokenClientConfig struct {
	VaultAddr string
	Token     string
}

type TokenClientParams struct {
	fx.In
	Lc  fx.Lifecycle
	Cfg *TokenClientConfig
}

func NewTokenClient(p TokenClientParams) (Client, error) {
	cfg := vault.DefaultConfig()
	if cfg.Error != nil {
		return nil, cfg.Error
	}

	cfg.Address = p.Cfg.VaultAddr

	vc, err := vault.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	vc.SetToken(p.Cfg.Token)

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

func newClient(lc fx.Lifecycle, addr string, auth vault.AuthMethod) (Client, error) {
	cfg := vault.DefaultConfig()
	if cfg.Error != nil {
		return nil, cfg.Error
	}

	cfg.Address = addr

	vc, err := vault.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	c := client{vc: vc, auth: auth}

	lc.Append(fx.Hook{
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
