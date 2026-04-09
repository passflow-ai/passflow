package main

import (
	"github.com/passflow-ai/passflow/cmd/passflow-channels/domain"
	"github.com/passflow-ai/passflow/cmd/passflow-channels/output"
)

func mergeRuleForUpdate(incoming domain.TriggerRule, existing *domain.TriggerRule) domain.TriggerRule {
	if existing == nil {
		return incoming
	}

	if incoming.CreatedAt.IsZero() {
		incoming.CreatedAt = existing.CreatedAt
	}

	if incoming.Auth == nil {
		incoming.Auth = existing.Auth
	} else if existing.Auth != nil && incoming.Auth.Secret == "" {
		incoming.Auth.Secret = existing.Auth.Secret
	}

	if incoming.Action.OutputChannel != nil {
		incoming.Action.OutputChannel = output.MergeSensitiveOutputChannel(existing.Action.OutputChannel, incoming.Action.OutputChannel)
	}

	if incoming.WebhookSecret == "" {
		incoming.WebhookSecret = existing.WebhookSecret
	}

	return incoming
}
