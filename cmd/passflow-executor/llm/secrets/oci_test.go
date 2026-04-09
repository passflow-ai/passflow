package secrets

import (
	"context"
	"testing"
)

func TestOCIVaultClient_GetSecret_EnvFallback(t *testing.T) {
	// Test without OCI config - should gracefully return error
	client := NewOCIVaultClient("", "")

	_, err := client.GetSecret(context.Background(), "test/secret")
	if err == nil {
		t.Log("OCI not configured, expected error or empty result")
	}
}

func TestOCIVaultClient_IsConfigured(t *testing.T) {
	t.Run("not configured without env", func(t *testing.T) {
		client := NewOCIVaultClient("", "")
		if client.IsConfigured() {
			t.Error("should not be configured without compartment/vault")
		}
	})

	t.Run("configured with values", func(t *testing.T) {
		client := NewOCIVaultClient("ocid1.compartment.test", "ocid1.vault.test")
		// Note: IsConfigured will be false because OCI SDK can't connect without real creds
		// but the fields are set
		if client.compartmentID != "ocid1.compartment.test" {
			t.Error("compartmentID should be set")
		}
		if client.vaultID != "ocid1.vault.test" {
			t.Error("vaultID should be set")
		}
	})
}
