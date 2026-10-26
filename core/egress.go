package core

import (
	"github.com/metacubex/mihomo/component/dialer"
	"github.com/metacubex/mihomo/config"
	"github.com/metacubex/mihomo/log"
)

// See "Egress Bootstrap" in .agents/architecture.md.
var (
	egressInterface string
	detectEgress    = func() string { return "" }
)

func prepareEgress(cfg *config.Config) *config.Config {
	egressInterface = ""
	general := cfg.General
	if general == nil ||
		general.Interface != "" ||
		!general.Tun.Enable ||
		!general.Tun.AutoRoute ||
		!general.Tun.AutoDetectInterface ||
		dialer.DefaultInterfaceFinder.Load() != nil {
		return cfg
	}
	name := detectEgress()
	if name == "" || name == general.Tun.Device {
		return cfg
	}
	pinned := *general
	pinned.Interface = name
	applied := *cfg
	applied.General = &pinned
	egressInterface = name
	log.Infoln("[TUN] bootstrap egress interface %s until auto-detect is up", name)
	return &applied
}

func finishEgress(cfg *config.Config) {
	if egressInterface == "" {
		return
	}
	configured := ""
	if cfg != nil && cfg.General != nil {
		configured = cfg.General.Interface
	}
	// A value updateConfig wrote meanwhile is left alone.
	dialer.DefaultInterface.CompareAndSwap(egressInterface, configured)
	egressInterface = ""
	log.Infoln("[TUN] bootstrap egress released, auto-detect finder up: %t", dialer.DefaultInterfaceFinder.Load() != nil)
}
