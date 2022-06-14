/* Copyright 2022 Zinc Labs Inc. and Contributors
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/pyroscope-io/client/pyroscope"
	"github.com/rs/zerolog/log"

	"github.com/zinclabs/zinc/pkg/config"
	"github.com/zinclabs/zinc/pkg/core"
	"github.com/zinclabs/zinc/pkg/meta"
	"github.com/zinclabs/zinc/pkg/metadata"
	"github.com/zinclabs/zinc/pkg/routes"
)

// @title           Zinc Search engine API
// @version         1.0
// @description     Zinc Search engine API documents https://docs.zincsearch.com
// @termsOfService  http://swagger.io/terms/

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @contact.name   Zinc Search
// @contact.url    https://www.zincsearch.com

// @securityDefinitions.basic BasicAuth
// @host      localhost:4080
// @BasePath  /
func main() {
	// Version
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("zinc version %s\n", meta.Version)
		os.Exit(0)
	}
	log.Info().Msgf("Starting Zinc %s", meta.Version)

	// Initialize sentry
	sentries()

	// Coninuous profiling
	profiling()

	// Index init
	core.LoadIndexList()

	// HTTP init
	app := gin.New()
	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	app.Use(gin.Recovery())
	// Debug for gin
	if gin.Mode() == gin.DebugMode {
		routes.AccessLog(app)
		routes.SetPProf(app)
	}
	routes.SetPrometheus(app) // Set up Prometheus.
	routes.SetRoutes(app)     // Set up all API routes.

	// Run the server
	PORT := config.Global.ServerPort
	server := &http.Server{
		Addr:    ":" + PORT,
		Handler: app,
	}
	shutdown(func(grace bool) {
		// close indexes
		err := core.ZINC_INDEX_LIST.Close()
		log.Info().Err(err).Msgf("Index closed")
		// close metadata
		err = metadata.Close()
		log.Info().Err(err).Msgf("Metadata closed")
		if grace {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				log.Fatal().Err(err).Msg("Server Shutdown")
			}
		} else {
			server.Close()
		}
	})

	if err := server.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			log.Info().Msg("Server closed under request")
		} else {
			log.Fatal().Err(err).Msg("Server closed unexpect")
		}
	}
	log.Info().Msg("Server shutdown ok")
}

func sentries() {
	if !config.Global.SentryEnable {
		return
	}
	if config.Global.SentryDSN == "" {
		return
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:     config.Global.SentryDSN,
		Release: "zinc@" + meta.Version,
	})
	if err != nil {
		log.Print("sentry.Init: ", err.Error())
	}
}

func profiling() {
	if !config.Global.ProfilerEnable {
		return
	}
	if config.Global.ProfilerServer == "" {
		return
	}

	ProfileID := config.Global.ProfilerFriendlyProfileID
	if ProfileID == "" {
		ProfileID = strings.ToLower(core.Telemetry.GetInstanceID())
	}

	pyroscope.Start(pyroscope.Config{
		ApplicationName: "zincsearch-" + ProfileID,

		// replace this with the address of pyroscope server
		ServerAddress: config.Global.ProfilerServer,

		// you can disable logging by setting this to nil
		// Logger: pyroscope.StandardLogger,
		Logger: nil,

		// optionally, if authentication is enabled, specify the API key:
		// AuthToken: os.Getenv("PYROSCOPE_AUTH_TOKEN"),
		AuthToken: config.Global.ProfilerAPIKey,

		// by default all profilers are enabled,
		// but you can select the ones you want to use:
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
		},
	})
}

//shutdown support twice signal must exit
func shutdown(stop func(grace bool)) {
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, syscall.SIGQUIT, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-sig
		go stop(s != syscall.SIGQUIT)
		<-sig
		os.Exit(128 + int(s.(syscall.Signal))) // second signal. Exit directly.
	}()
}
