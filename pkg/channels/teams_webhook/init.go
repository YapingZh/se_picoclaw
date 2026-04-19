package teamswebhook

import (
	"study/picoclaw/pkg/bus"
	"study/picoclaw/pkg/channels"
	"study/picoclaw/pkg/config"
)

func init() {
	channels.RegisterFactory("teams_webhook", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewTeamsWebhookChannel(cfg.Channels.TeamsWebhook, b)
	})
}
