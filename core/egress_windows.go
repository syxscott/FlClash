//go:build windows

package core

import (
	"github.com/metacubex/mihomo/listener/sing_tun"
	"github.com/metacubex/mihomo/log"
	tun "github.com/metacubex/sing-tun"
)

func init() {
	detectEgress = detectWindowsEgress
}

// The interface a fresh connection would leave through, computed the way
// sing-tun's own monitor picks the auto-detect answer: the lowest metric
// default route on an up, connected, non-virtual interface. The monitors are
// built and torn down for this one read; the TUN listener owns the long-lived
// pair once it starts.
func detectWindowsEgress() string {
	network, err := tun.NewNetworkUpdateMonitor(log.SingLogger)
	if err != nil {
		return ""
	}
	defer network.Close()
	monitor, err := tun.NewDefaultInterfaceMonitor(network, log.SingLogger, tun.DefaultInterfaceMonitorOptions{
		InterfaceFinder: sing_tun.DefaultInterfaceFinder,
	})
	if err != nil {
		return ""
	}
	defer monitor.Close()
	if err := monitor.Start(); err != nil {
		return ""
	}
	if selected := monitor.DefaultInterface(); selected != nil {
		return selected.Name
	}
	return ""
}
