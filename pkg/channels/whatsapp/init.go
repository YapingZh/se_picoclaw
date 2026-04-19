package whatsapp

import (
	"study/picoclaw/pkg/bus"
	"study/picoclaw/pkg/channels"
	"study/picoclaw/pkg/config"
)

func init() {
	channels.RegisterFactory("whatsapp", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewWhatsAppChannel(cfg.Channels.WhatsApp, b)
	})
}
