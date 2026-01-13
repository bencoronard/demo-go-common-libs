package vault

import (
	"context"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/fx"

	authAppRole "github.com/hashicorp/vault/api/auth/approle"
	authK8s "github.com/hashicorp/vault/api/auth/kubernetes"
	authUsrPsw "github.com/hashicorp/vault/api/auth/userpass"
)

type Client interface {
	ReadSecret(ctx context.Context, path string, target any) error
	WatchTokenLifecycle(lc fx.Lifecycle, sd fx.Shutdowner) error
}

type client struct {
	vc   *vault.Client
	auth vault.AuthMethod
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

func (c *client) WatchTokenLifecycle(lc fx.Lifecycle, sd fx.Shutdowner) error {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go c.renewToken(ctx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
	return nil
}

func (c *client) renewToken(ctx context.Context) {
	for {
		if err := c.manageTokenLifecycle(); err != nil {
			log.Fatalf("unable to start managing token lifecycle: %v", err)
		}

		_, err := c.login(ctx)
		if err != nil {
			log.Fatalf("unable to authenticate to Vault: %v", err)
		}
	}
}

func (c *client) manageTokenLifecycle() error {
	token, err := c.vc.Auth().Token().LookupSelf()
	if err != nil {
		return err
	}

	if !token.Auth.Renewable {
		log.Printf("Token is not renewable. Re-attempting login")
		return nil
	}

	watcher, err := c.vc.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
		Secret:    token,
		Increment: 3600,
	})
	if err != nil {
		return err
	}

	go watcher.Start()
	defer watcher.Stop()

	for {
		select {
		case err := <-watcher.DoneCh():
			if err != nil {
				log.Printf("Failed to renew token: %v. Re-attempting login", err)
				return nil
			}
			log.Printf("Token can no longer be renewed. Re-attempting login.")
			return nil
		case renewal := <-watcher.RenewCh():
			log.Printf("Successfully renewed: %#v", renewal)
		}
	}
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
			_, err := c.login(ctx)
			if err != nil {
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

func NewAppRoleClient(lc fx.Lifecycle, addr, roleID, secretID string) (Client, error) {
	cfg := vault.DefaultConfig()
	cfg.Address = addr

	vc, err := vault.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	auth, err := authAppRole.NewAppRoleAuth(roleID, &authAppRole.SecretID{FromString: secretID})
	if err != nil {
		return nil, err
	}

	c := client{vc: vc, auth: auth}

	return &c, nil
}

func NewUserPassClient(lc fx.Lifecycle, addr, usr, psw string) (Client, error) {
	cfg := vault.DefaultConfig()
	cfg.Address = addr

	vc, err := vault.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	auth, err := authUsrPsw.NewUserpassAuth(usr, &authUsrPsw.Password{FromString: psw})
	if err != nil {
		return nil, err
	}

	c := client{vc: vc, auth: auth}

	return &c, nil
}
