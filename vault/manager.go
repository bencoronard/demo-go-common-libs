package vault

import (
	"context"
	"log"

	vault "github.com/hashicorp/vault/api"
)

func (c *client) renewToken(ctx context.Context) {
	for {
		if err := c.manageTokenLifecycle(); err != nil {
			log.Fatalf("unable to start managing token lifecycle: %v", err)
		}

		_, err := c.authenticate(ctx)
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
