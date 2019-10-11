/*
 * FTP Server Go
 *
 * Copyright (C) 2019 Donam Kim. All rights reserved.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along
 * with this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
 */

package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/donamKim/ftp-server-go/pi"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().Unix())
}

func main() {
	initConfig()
	initServer()
	waitSignal()
}

func initConfig() {
	viper.SetConfigFile("/usr/local/etc/ftp/server.yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read the config file: %v", err)
	}
}

func initServer() {
	svr := pi.Server{
		User: struct {
			Name     string
			Password string
		}{
			Name:     viper.GetString("user.name"),
			Password: viper.GetString("user.password"),
		},
		Root:        viper.GetString("root"),
		PIPort:      viper.GetInt("pi_port"),
		PassivePort: cast.ToIntSlice(viper.Get("passive_port")),
	}
	svr.ListenAndServe()
}

func waitSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGPIPE)

	for {
		s := <-c
		switch s {
		case syscall.SIGTERM, syscall.SIGINT:
			log.Fatalf("caught %v signal: shutting down...", s)
			return
		default:
			log.Fatalf("caught %v signal: ignored!", s)
		}
	}
}
