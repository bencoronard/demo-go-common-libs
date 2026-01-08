package vault

import (
	"context"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/fx"

	authAppRole "github.com/hashicorp/vault/api/auth/approle"
	authK8s "github.com/hashicorp/vault/api/auth/kubernetes"
	authUsrPsw "github.com/hashicorp/vault/api/auth/userpass"
)

type VaultClient interface {
	GetSecret(ctx context.Context, path string, target any) error
}

type client struct {
	vault *vault.Client
}

func NewVaultTokenClient(lc fx.Lifecycle, addr, token string) (VaultClient, error) {
	vc, err := initClient(addr)
	if err != nil {
		return nil, err
	}

	vc.SetToken(token)

	return &client{vault: vc}, nil
}

func NewVaultUserPassClient(lc fx.Lifecycle, addr, usr, psw string) (VaultClient, error) {
	vc, err := initClient(addr)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return authWithUserPass(ctx, vc, usr, psw)
		},
	})

	return &client{vault: vc}, nil
}

func NewVaultAppRoleClient(lc fx.Lifecycle, addr, roleID, secretID string) (VaultClient, error) {
	vc, err := initClient(addr)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return authWithAppRole(ctx, vc, roleID, secretID)
		},
	})

	return &client{vault: vc}, nil
}

func NewVaultK8sClient(lc fx.Lifecycle, addr, role, token string) (VaultClient, error) {
	vc, err := initClient(addr)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return authWithK8s(ctx, vc, role, token)
		},
	})

	return &client{vault: vc}, nil
}

func (c *client) GetSecret(ctx context.Context, path string, target any) error {
	secret, err := c.vault.Logical().ReadWithContext(ctx, path)
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

func initClient(addr string) (*vault.Client, error) {
	cfg := vault.DefaultConfig()
	cfg.Address = addr

	client, err := vault.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func authWithUserPass(ctx context.Context, vc *vault.Client, usr, psw string) error {
	auth, err := authUsrPsw.NewUserpassAuth(usr, &authUsrPsw.Password{FromString: psw})
	if err != nil {
		return err
	}

	authInfo, err := vc.Auth().Login(ctx, auth)
	if err != nil {
		return err
	}
	if authInfo == nil {
		return fmt.Errorf("%w: no auth info returned from UserPass login", ErrAuthenticationFail)
	}

	return nil
}

func authWithAppRole(ctx context.Context, vc *vault.Client, roleID, secretID string) error {
	auth, err := authAppRole.NewAppRoleAuth(roleID, &authAppRole.SecretID{FromString: secretID})
	if err != nil {
		return err
	}

	authInfo, err := vc.Auth().Login(ctx, auth)
	if err != nil {
		return err
	}
	if authInfo == nil {
		return fmt.Errorf("%w: no auth info returned from AppRole login", ErrAuthenticationFail)
	}

	return nil
}

func authWithK8s(ctx context.Context, vc *vault.Client, role, token string) error {
	auth, err := authK8s.NewKubernetesAuth(role, authK8s.WithServiceAccountToken(token))
	if err != nil {
		return err
	}

	authInfo, err := vc.Auth().Login(ctx, auth)
	if err != nil {
		return err
	}
	if authInfo == nil {
		return fmt.Errorf("%w: no auth info returned from Kubernetes login", ErrAuthenticationFail)
	}

	return nil
}
