package vk

import (
	"study/picoclaw/pkg/bus"
	"study/picoclaw/pkg/channels"
	"study/picoclaw/pkg/config"
)

func init() {
	channels.RegisterFactory("vk", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewVKChannel(cfg, b)
	})
}
