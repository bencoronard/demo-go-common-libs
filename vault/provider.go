package vault

import (
	"context"
	"os"

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

type Params struct {
	fx.In
	Lc fx.Lifecycle
}

func NewK8sClient(p Params) (Client, error) {
	role := os.Getenv("VAULT_ROLE")
	if role == "" {
		return nil, ErrConfigUnset
	}

	auth, err := authK8s.NewKubernetesAuth(role)
	if err != nil {
		return nil, err
	}

	return newClient(p.Lc, auth)
}

func NewAppRoleClient(p Params) (Client, error) {
	role := os.Getenv("VAULT_ROLE")
	if role == "" {
		return nil, ErrConfigUnset
	}

	auth, err := authAppRole.NewAppRoleAuth(role, &authAppRole.SecretID{FromFile: "VAULT_SECRET"})
	if err != nil {
		return nil, err
	}

	return newClient(p.Lc, auth)
}

func NewUserPassClient(p Params) (Client, error) {
	user := os.Getenv("VAULT_USER")
	if user == "" {
		return nil, ErrConfigUnset
	}

	auth, err := authUsrPsw.NewUserpassAuth(user, &authUsrPsw.Password{FromEnv: "VAULT_SECRET"})
	if err != nil {
		return nil, err
	}
	return newClient(p.Lc, auth)
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

func newClient(lc fx.Lifecycle, auth vault.AuthMethod) (Client, error) {
	cfg := vault.DefaultConfig()
	if cfg.Error != nil {
		return nil, cfg.Error
	}

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
