package vault

import (
	"context"

	vault "github.com/hashicorp/vault/api"
	"go.uber.org/fx"
)

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
