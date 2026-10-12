package main

import (
	"log"
	"sync"

	"github.com/godbus/dbus/v5"
)

const (
	stateUnknown  = uint32(0)
	stateInit     = uint32(1)
	stateShutdown = uint32(2)
	stateStarting = uint32(3)
	stateStarted  = uint32(4)
	stateStopping = uint32(5)
	stateStopped  = uint32(6)

	failureLoginFailed   = uint32(0)
	failureConnectFailed = uint32(1)

	pluginPath  = dbus.ObjectPath("/org/freedesktop/NetworkManager/VPN/Plugin")
	pluginIface = "org.freedesktop.NetworkManager.VPN.Plugin"
)

type Plugin struct {
	conn  *dbus.Conn
	quit  chan struct{}
	state uint32
	mu    sync.Mutex
}

func NewPlugin(conn *dbus.Conn, quit chan struct{}) *Plugin {
	return &Plugin{conn: conn, quit: quit, state: stateStopped}
}

// --- D-Bus Properties interface (NM reads State as a property) ---

type pluginProperties struct{ p *Plugin }

func (pp *pluginProperties) Get(iface, prop string) (dbus.Variant, *dbus.Error) {
	if iface == pluginIface && prop == "State" {
		pp.p.mu.Lock()
		defer pp.p.mu.Unlock()
		return dbus.MakeVariant(pp.p.state), nil
	}
	return dbus.Variant{}, dbus.NewError("org.freedesktop.DBus.Error.InvalidArgs", nil)
}

func (pp *pluginProperties) GetAll(iface string) (map[string]dbus.Variant, *dbus.Error) {
	if iface == pluginIface {
		pp.p.mu.Lock()
		defer pp.p.mu.Unlock()
		return map[string]dbus.Variant{"State": dbus.MakeVariant(pp.p.state)}, nil
	}
	return nil, dbus.NewError("org.freedesktop.DBus.Error.InvalidArgs", nil)
}

func (pp *pluginProperties) Set(_, _ string, _ dbus.Variant) *dbus.Error {
	return dbus.NewError("org.freedesktop.DBus.Error.NotSupported", nil)
}

// --- internal helpers ---

func (p *Plugin) setState(state uint32) {
	p.mu.Lock()
	p.state = state
	p.mu.Unlock()
	p.conn.Emit(pluginPath, pluginIface+".StateChanged", state) //nolint:errcheck
}

func (p *Plugin) emitFailure(reason uint32) {
	p.conn.Emit(pluginPath, pluginIface+".Failure", reason) //nolint:errcheck
	p.setState(stateStopped)
	p.quit <- struct{}{}
}

func (p *Plugin) emitConfig() {
	tundev := "wt0"
	if s, err := getNetbirdStatus(); err == nil && s.TunIface != "" {
		tundev = s.TunIface
	}
	cfg := map[string]dbus.Variant{
		"tundev":        dbus.MakeVariant(tundev),
		"has-ip4":       dbus.MakeVariant(false),
		"has-ip6":       dbus.MakeVariant(false),
		"never-default": dbus.MakeVariant(true),
	}
	p.conn.Emit(pluginPath, pluginIface+".Config", cfg) //nolint:errcheck
}

func (p *Plugin) connectAsync() {
	err := runNetbird("up")
	if err != nil {
		if isAuthError(err) {
			log.Println("auth required — opening browser login")
			p.conn.Emit(pluginPath, pluginIface+".LoginBanner", //nolint:errcheck
				"Waiting for browser authentication…")
			if loginErr := runNetbird("login"); loginErr != nil {
				log.Printf("netbird login failed: %v", loginErr)
				p.emitFailure(failureLoginFailed)
				return
			}
			err = runNetbird("up")
		}
		if err != nil {
			log.Printf("netbird up failed: %v", err)
			p.emitFailure(failureConnectFailed)
			return
		}
	}
	p.emitConfig()
	p.setState(stateStarted)
}

// --- NM VPN Plugin interface methods ---

func (p *Plugin) Connect(connection map[string]map[string]dbus.Variant) *dbus.Error {
	p.setState(stateStarting)
	go p.connectAsync()
	return nil
}

func (p *Plugin) ConnectInteractive(connection map[string]map[string]dbus.Variant, _ map[string]dbus.Variant) *dbus.Error {
	return p.Connect(connection)
}

func (p *Plugin) NeedSecrets(_ map[string]map[string]dbus.Variant) (string, *dbus.Error) {
	return "", nil
}

func (p *Plugin) Disconnect() *dbus.Error {
	p.setState(stateStopping)
	go func() {
		if err := runNetbird("down"); err != nil {
			log.Printf("netbird down: %v", err)
		}
		p.setState(stateStopped)
		p.quit <- struct{}{}
	}()
	return nil
}

func (p *Plugin) GetState() (uint32, *dbus.Error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state, nil
}

// Stub methods NM may call but we don't need.
func (p *Plugin) SetConfig(_ map[string]dbus.Variant) *dbus.Error          { return nil }
func (p *Plugin) SetIp4Config(_ map[string]dbus.Variant) *dbus.Error       { return nil }
func (p *Plugin) SetIp6Config(_ map[string]dbus.Variant) *dbus.Error       { return nil }
func (p *Plugin) SetFailure(_ string) *dbus.Error                          { return nil }
func (p *Plugin) NewSecrets(_ map[string]map[string]dbus.Variant) *dbus.Error { return nil }
