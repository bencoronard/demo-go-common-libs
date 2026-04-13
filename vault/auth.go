package vault

import (
	"os"
	"strings"

	"github.com/hashicorp/vault/api/auth/approle"
	"github.com/hashicorp/vault/api/auth/kubernetes"
	"github.com/hashicorp/vault/api/auth/userpass"
)

func NewK8sAuth() (*kubernetes.KubernetesAuth, error) {
	role := strings.TrimSpace(os.Getenv("VAULT_K8S_ROLE"))
	if role == "" {
		return nil, ErrConfigUnset
	}

	auth, err := kubernetes.NewKubernetesAuth(role)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func NewAppRoleAuth() (*approle.AppRoleAuth, error) {
	role := strings.TrimSpace(os.Getenv("VAULT_APP_ROLE"))
	if role == "" {
		return nil, ErrConfigUnset
	}

	auth, err := approle.NewAppRoleAuth(
		role,
		&approle.SecretID{FromFile: ""},
	)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func NewUserPassAuth() (*userpass.UserpassAuth, error) {
	user := strings.TrimSpace(os.Getenv("VAULT_USER"))
	if user == "" {
		return nil, ErrConfigUnset
	}

	auth, err := userpass.NewUserpassAuth(
		user,
		&userpass.Password{FromEnv: ""},
	)
	if err != nil {
		return nil, err
	}

	return auth, nil
}
