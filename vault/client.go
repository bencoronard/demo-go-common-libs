package vault

import (
	"context"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/fx"

	auth "github.com/hashicorp/vault/api/auth/kubernetes"
)

type VaultClient interface {
	GetSecret(ctx context.Context, path string, target any) error
}

type client struct {
	vault *vault.Client
}

func NewVaultTokenClient(lc fx.Lifecycle, addr, token string) (VaultClient, error) {
	cfg := vault.DefaultConfig()
	cfg.Address = addr

	vc, err := vault.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	vc.SetToken(token)

	return &client{vault: vc}, nil
}

func NewVaultAppRoleClient(lc fx.Lifecycle, addr, roleID, secretID string) (VaultClient, error) {
	cfg := vault.DefaultConfig()
	cfg.Address = addr

	vc, err := vault.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	data := map[string]any{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	resp, err := vc.Logical().WriteWithContext(context.Background(), "auth/approle/login", data)
	if err != nil {
		return nil, err
	}

	vc.SetToken(resp.Auth.ClientToken)

	return &client{vault: vc}, nil
}

func NewVaultK8sClient(lc fx.Lifecycle, addr, role, tokenPath string) (VaultClient, error) {
	cfg := vault.DefaultConfig()
	cfg.Address = addr

	vc, err := vault.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	k8sAuth, err := auth.NewKubernetesAuth(role, auth.WithMountPath(tokenPath))
	if err != nil {
		return nil, err
	}

	authInfo, err := vc.Auth().Login(context.Background(), k8sAuth)
	if err != nil {
		return nil, err
	}

	if authInfo == nil {
		return nil, fmt.Errorf("%w: no auth info returned from kubernetes login", ErrAuthenticationFail)
	}

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
