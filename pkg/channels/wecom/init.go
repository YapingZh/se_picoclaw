package wecom

import (
	"study/picoclaw/pkg/bus"
	"study/picoclaw/pkg/channels"
	"study/picoclaw/pkg/config"
)

func init() {
	channels.RegisterFactory("wecom", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewChannel(cfg.Channels.WeCom, b)
	})
}
