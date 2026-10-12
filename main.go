package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/godbus/dbus/v5"
)

const busName = "org.freedesktop.NetworkManager.netbird"

func main() {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		log.Fatalf("connect system bus: %v", err)
	}
	defer conn.Close()

	reply, err := conn.RequestName(busName, dbus.NameFlagDoNotQueue)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		log.Fatalf("acquire bus name %q: %v (reply=%d)", busName, err, reply)
	}

	quit := make(chan struct{}, 1)
	plugin := NewPlugin(conn, quit)

	conn.Export(plugin, pluginPath, pluginIface)
	conn.Export(&pluginProperties{plugin}, pluginPath, "org.freedesktop.DBus.Properties")

	log.Printf("nm-netbird-service ready on %s", busName)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-quit:
		log.Println("connection lifecycle complete, exiting")
	case <-sig:
		log.Println("signal received, exiting")
	}
}
