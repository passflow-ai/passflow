package secrets

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/secrets"
)

// OCIVaultClient retrieves secrets from OCI Vault.
type OCIVaultClient struct {
	compartmentID string
	vaultID       string
	client        *secrets.SecretsClient
	configured    bool
}

// NewOCIVaultClient creates a new OCI Vault client.
// If compartmentID or vaultID are empty, the client is not configured.
func NewOCIVaultClient(compartmentID, vaultID string) *OCIVaultClient {
	c := &OCIVaultClient{
		compartmentID: compartmentID,
		vaultID:       vaultID,
		configured:    compartmentID != "" && vaultID != "",
	}

	if c.configured {
		provider := common.DefaultConfigProvider()
		client, err := secrets.NewSecretsClientWithConfigurationProvider(provider)
		if err == nil {
			c.client = &client
		} else {
			c.configured = false
		}
	}

	return c
}

// IsConfigured returns true if OCI Vault is configured.
func (c *OCIVaultClient) IsConfigured() bool {
	return c.configured
}

// GetSecret retrieves a secret by name from OCI Vault.
func (c *OCIVaultClient) GetSecret(ctx context.Context, secretName string) (string, error) {
	if !c.configured || c.client == nil {
		return "", fmt.Errorf("oci vault not configured")
	}

	req := secrets.GetSecretBundleByNameRequest{
		SecretName: common.String(secretName),
		VaultId:    common.String(c.vaultID),
	}

	resp, err := c.client.GetSecretBundleByName(ctx, req)
	if err != nil {
		return "", fmt.Errorf("oci vault: failed to get secret %q: %w", secretName, err)
	}

	content, ok := resp.SecretBundleContent.(secrets.Base64SecretBundleContentDetails)
	if !ok {
		return "", fmt.Errorf("oci vault: unexpected content type for %q", secretName)
	}

	decoded, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return "", fmt.Errorf("oci vault: failed to decode secret %q: %w", secretName, err)
	}

	return string(decoded), nil
}
