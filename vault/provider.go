package vault

import (
	"context"

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

type ClientParams struct {
	fx.In
	Lc fx.Lifecycle
}

type K8sClientConfig struct {
	VaultAddr      string
	RoleName       string
	TokenMountPath string
}

type K8sClientParams struct {
	ClientParams
	Cfg K8sClientConfig
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
	ClientParams
	Cfg AppRoleClientConfig
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
	ClientParams
	Cfg UserPassClientConfig
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
	ClientParams
	Cfg TokenClientConfig
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
