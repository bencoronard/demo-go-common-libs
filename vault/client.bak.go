package vault

import (
	"context"
	"fmt"

	vault "github.com/hashicorp/vault/api"

	authAppRole "github.com/hashicorp/vault/api/auth/approle"
	authK8s "github.com/hashicorp/vault/api/auth/kubernetes"
	authUsrPsw "github.com/hashicorp/vault/api/auth/userpass"
)

// func (c *client) manageLifecycle(lc fx.Lifecycle, initialSecret *vault.Secret, auth authFunc) {
// 	lc.Append(fx.Hook{
// 		OnStart: func(ctx context.Context) error {
// 			go c.runRenewalLoop(initialSecret, auth)
// 			return nil
// 		},
// 	})
// }

// func NewVaultTokenClient(lc fx.Lifecycle, addr, token string) (VaultClient, error) {
// 	vc, err := initClient(addr)
// 	if err != nil {
// 		return nil, err
// 	}

// 	vc.SetToken(token)

// 	return &client{vault: vc}, nil
// }

// func NewVaultUserPassClient(lc fx.Lifecycle, addr, usr, psw string) (VaultClient, error) {
// 	vc, err := initClient(addr)
// 	if err != nil {
// 		return nil, err
// 	}

// 	lc.Append(fx.Hook{
// 		OnStart: func(ctx context.Context) error {
// 			return authWithUserPass(ctx, vc, usr, psw)
// 		},
// 	})

// 	return &client{vault: vc}, nil
// }

// func NewVaultAppRoleClient(lc fx.Lifecycle, addr, roleID, secretID string) (VaultClient, error) {
// 	vc, err := initClient(addr)
// 	if err != nil {
// 		return nil, err
// 	}

// 	lc.Append(fx.Hook{
// 		OnStart: func(ctx context.Context) error {
// 			return authWithAppRole(ctx, vc, roleID, secretID)
// 		},
// 	})

// 	return &client{vault: vc}, nil
// }

// func NewVaultK8sClient(lc fx.Lifecycle, addr, role, token string) (VaultClient, error) {
// 	vc, err := initClient(addr)
// 	if err != nil {
// 		return nil, err
// 	}

// 	lc.Append(fx.Hook{
// 		OnStart: func(ctx context.Context) error {
// 			return authWithK8s(ctx, vc, role, token)
// 		},
// 	})

// 	return &client{vault: vc}, nil
// }

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

// func (c *client) runRenewalLoop(initialSecret *vault.Secret, auth authFunc) {
// 	currentSecret := initialSecret

// 	for {
// 		watcher, err := c.vault.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
// 			Secret: currentSecret,
// 		})
// 		if err != nil {
// 			slog.Info(fmt.Sprintf("[Vault] Failed to create watcher: %v. Retrying login in 10s...", err))
// 			time.Sleep(10 * time.Second)
// 			goto reauth
// 		}

// 		go watcher.Start()

// 		select {
// 		case err := <-watcher.DoneCh():
// 			watcher.Stop()
// 			if err != nil {
// 				slog.Info(fmt.Sprintf("[Vault] Token expired or renewal error: %v. Re-authenticating...", err))
// 			} else {
// 				slog.Info("[Vault] Max TTL reached. Re-authenticating...")
// 			}
// 		case renewal := <-watcher.RenewCh():
// 			slog.Info(fmt.Sprintf("[Vault] Token renewed at %s", renewal.RenewedAt))
// 			continue
// 		}

// 	reauth:
// 		newSecret, err := auth(context.Background(), c.vault)
// 		if err != nil {
// 			slog.Info(fmt.Sprintf("[Vault] Re-auth failed: %v. Retrying in 30s...", err))
// 			time.Sleep(30 * time.Second)
// 			goto reauth
// 		}

// 		c.mu.Lock()
// 		c.vault.SetToken(newSecret.Auth.ClientToken)
// 		currentSecret = newSecret
// 		c.mu.Unlock()
// 	}
// }// func (c *client) runRenewalLoop(initialSecret *vault.Secret, auth authFunc) {
// 	currentSecret := initialSecret

// 	for {
// 		watcher, err := c.vault.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
// 			Secret: currentSecret,
// 		})
// 		if err != nil {
// 			slog.Info(fmt.Sprintf("[Vault] Failed to create watcher: %v. Retrying login in 10s...", err))
// 			time.Sleep(10 * time.Second)
// 			goto reauth
// 		}

// 		go watcher.Start()

// 		select {
// 		case err := <-watcher.DoneCh():
// 			watcher.Stop()
// 			if err != nil {
// 				slog.Info(fmt.Sprintf("[Vault] Token expired or renewal error: %v. Re-authenticating...", err))
// 			} else {
// 				slog.Info("[Vault] Max TTL reached. Re-authenticating...")
// 			}
// 		case renewal := <-watcher.RenewCh():
// 			slog.Info(fmt.Sprintf("[Vault] Token renewed at %s", renewal.RenewedAt))
// 			continue
// 		}

// 	reauth:
// 		newSecret, err := auth(context.Background(), c.vault)
// 		if err != nil {
// 			slog.Info(fmt.Sprintf("[Vault] Re-auth failed: %v. Retrying in 30s...", err))
// 			time.Sleep(30 * time.Second)
// 			goto reauth
// 		}

// 		c.mu.Lock()
// 		c.vault.SetToken(newSecret.Auth.ClientToken)
// 		currentSecret = newSecret
// 		c.mu.Unlock()
// 	}
// }
