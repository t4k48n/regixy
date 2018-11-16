package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/sys/windows/registry"
)

const InternetSettingsKey = `Software\Microsoft\Windows\CurrentVersion\Internet Settings`

func getInternetSettingsKey(access uint32) (registry.Key, error) {
	return registry.OpenKey(
		registry.CURRENT_USER,
		InternetSettingsKey,
		access,
	)
}

func GetEnable() (bool, error) {
	key, err := getInternetSettingsKey(registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer key.Close()
	val, _, err := key.GetIntegerValue("ProxyEnable")
	if err != nil {
		return false, err
	}
	switch val {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("invalid value read %v", val)
	}
}

func GetServer() (string, error) {
	key, err := getInternetSettingsKey(registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer key.Close()
	val, _, err := key.GetStringValue("ProxyServer")
	if err != nil {
		return "", err
	}
	return val, nil
}

func SetEnable(enable bool) error {
	key, err := getInternetSettingsKey(registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	var v uint32
	if enable {
		v = 1
	} else {
		v = 0
	}
	err = key.SetDWordValue("ProxyEnable", v)
	if err != nil {
		return err
	}
	return nil
}

func SetServer(server string) error {
	key, err := getInternetSettingsKey(registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	err = key.SetStringValue("ProxyServer", server)
	if err != nil {
		return err
	}
	return nil
}

func WriteHelp(w io.Writer) {
	text := `usage: %v [status|on|off|help]
`
	fmt.Fprintf(w, text, os.Args[0])
}

func WriteStatus(w io.Writer) error {
	enable, err := GetEnable()
	if err != nil {
		return nil
	}
	server, err := GetServer()
	if err != nil {
		return nil
	}
	text := `Enable: %v
Server: %v
`
	fmt.Fprintf(w, text, enable, server)
	return nil
}

var subcommandToFunction = map[string]func() {
	"status": func () {
		if err := WriteStatus(os.Stdout); err != nil {
			log.Fatal(err)
		}
	},
	"on": func() {
		if err := SetEnable(true); err != nil {
			log.Fatal(err)
		}
	},
	"off": func() {
		if err := SetEnable(false); err != nil {
			log.Fatal(err)
		}
	},
	"help": func() {
		WriteHelp(os.Stdout)
	},
}

func main() {
	if len(os.Args) != 2 {
		WriteHelp(os.Stderr)
		os.Exit(1)
	}
	f, ok := subcommandToFunction[os.Args[1]]
	if !ok {
		WriteHelp(os.Stderr)
		os.Exit(1)
	}
	f()
}
