package core

import (
	"github.com/metacubex/mihomo/component/dialer"
	"github.com/metacubex/mihomo/config"
	"github.com/metacubex/mihomo/log"
)

// The first apply of a core process runs before the TUN listener installs
// dialer.DefaultInterfaceFinder, and FlClash leaves interface-name empty so
// auto-detect owns the egress. A dial made in that window that is still
// resolving when sing-tun's auto-route lands goes out unbound, is captured by
// the new TUN default route, and loops back into the core, which reports it as
// a dialer loop; the client sees an outage that lasts out the TCP retries.
// Pin those first dials to the interface the OS would pick anyway and release
// the pin once the finder is live, which every later dial consults instead.
// Windows only, where the outage is reported (#2853); elsewhere the detector
// answers "" and nothing is pinned.
//
// Both entry points run under configMu, from applyConfig and the listener
// paths, which is the only guard egressInterface needs.

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
	// The stored config keeps the interface the profile asked for; only the
	// copy handed to this apply carries the pin, which executor.ApplyConfig
	// writes into dialer.DefaultInterface.
	pinned := *general
	pinned.Interface = name
	prepared := *cfg
	prepared.General = &pinned
	egressInterface = name
	log.Infoln("[TUN] bootstrap egress interface %s until the auto-detect finder is up", name)
	return &prepared
}

func finishEgress(cfg *config.Config) {
	if egressInterface == "" {
		return
	}
	configured := ""
	if cfg != nil && cfg.General != nil {
		configured = cfg.General.Interface
	}
	// CompareAndSwap leaves a value updateConfig wrote meanwhile alone.
	dialer.DefaultInterface.CompareAndSwap(egressInterface, configured)
	egressInterface = ""
	log.Infoln("[TUN] bootstrap egress released, auto-detect finder up: %t", dialer.DefaultInterfaceFinder.Load() != nil)
}
