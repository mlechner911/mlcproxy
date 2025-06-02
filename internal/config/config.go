/*
Copyright (c) 2025 Michael Lechner

This software is released under the MIT License.
See the LICENSE file for further details.
*/

package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

type Config struct {
	Server struct {
		Port int
	}
	Paths struct {
		StaticDir string
		StatsPath string
		APIPath   string
	}
	Features struct {
		StatsHost string
	}
	Auth struct {
		EnableAuth  bool
		Credentials map[string]string // username -> password
	}
	Security struct {
		AllowedNetworks []string
	}
}

var Cfg Config

// LoadConfig lädt die Konfiguration aus der config.ini Datei
func LoadConfig() error {
	// Finde den Basispfad der Anwendung
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	basePath := filepath.Dir(executable)

	// Suche nach config.ini in verschiedenen Pfaden
	configPaths := []string{
		"config.ini",                          // Aktuelles Verzeichnis
		filepath.Join(basePath, "config.ini"), // Executable-Verzeichnis
		"../config.ini",                       // Ein Verzeichnis höher
		"../../config.ini",                    // Zwei Verzeichnisse höher
	}

	var cfg *ini.File
	var loadErr error
	for _, path := range configPaths {
		cfg, err = ini.Load(path)
		if err == nil {
			log.Printf("Konfiguration geladen aus: %s", path)
			break
		}
		loadErr = err
	}
	if loadErr != nil {
		return loadErr
	}

	// Server-Sektion
	Cfg.Server.Port = cfg.Section("server").Key("port").MustInt(3128)

	// Paths-Sektion mit absoluten Pfaden
	Cfg.Paths.StaticDir = filepath.Join(basePath, cfg.Section("paths").Key("static_dir").MustString("static"))
	Cfg.Paths.StatsPath = cfg.Section("paths").Key("stats_path").MustString("/stat")
	Cfg.Paths.APIPath = cfg.Section("paths").Key("api_path").MustString("/api")

	// Features-Sektion
	Cfg.Features.StatsHost = cfg.Section("features").Key("stats_host").MustString("stats.local")

	// Auth-Sektion
	authSec := cfg.Section("auth")
	Cfg.Auth.EnableAuth = authSec.Key("enable_auth").MustBool(false)

	// Credentials Map initialisieren
	Cfg.Auth.Credentials = make(map[string]string)
	if credStr := authSec.Key("credentials").String(); credStr != "" {
		// Verarbeite "user1:pass1,user2:pass2" Format
		for _, cred := range strings.Split(credStr, ",") {
			parts := strings.SplitN(cred, ":", 2)
			if len(parts) == 2 {
				Cfg.Auth.Credentials[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Security-Sektion
	secSec := cfg.Section("security")
	if networks := secSec.Key("allowed_networks").String(); networks != "" {
		Cfg.Security.AllowedNetworks = strings.Split(networks, ",")
		// Trimme Whitespace
		for i := range Cfg.Security.AllowedNetworks {
			Cfg.Security.AllowedNetworks[i] = strings.TrimSpace(Cfg.Security.AllowedNetworks[i])
		}
	} else {
		// Standard: Nur localhost
		Cfg.Security.AllowedNetworks = []string{"127.0.0.1/32"}
	}

	return nil
}
