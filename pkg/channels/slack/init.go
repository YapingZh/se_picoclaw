package slack

import (
	"study/picoclaw/pkg/bus"
	"study/picoclaw/pkg/channels"
	"study/picoclaw/pkg/config"
)

func init() {
	channels.RegisterFactory("slack", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewSlackChannel(cfg.Channels.Slack, b)
	})
}
