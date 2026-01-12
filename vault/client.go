package vault

import (
	"context"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/fx"

	authK8s "github.com/hashicorp/vault/api/auth/kubernetes"
)

type Client interface {
	ReadSecret(ctx context.Context, path string, target any) error
	authenticate(ctx context.Context) error
	client() *vault.Client
}

type client struct {
	vc   *vault.Client
	auth vault.AuthMethod
}

func (c *client) client() *vault.Client {
	return c.vc
}

func (c *client) authenticate(ctx context.Context) error {
	authInfo, err := c.vc.Auth().Login(ctx, c.auth)
	if err != nil {
		return err
	}
	if authInfo == nil {
		return fmt.Errorf("%w: no auth info returned", ErrAuthenticationFail)
	}
	return nil
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

func NewK8sClient(lc fx.Lifecycle, addr, role, mountPath string) (Client, error) {
	cfg := vault.DefaultConfig()
	cfg.Address = addr

	vc, err := vault.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	auth, err := authK8s.NewKubernetesAuth(role, authK8s.WithServiceAccountTokenPath(mountPath))
	if err != nil {
		return nil, err
	}

	c := client{vc: vc, auth: auth}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return c.authenticate(ctx)
		},
	})

	return &c, nil
}
