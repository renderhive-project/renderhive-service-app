/*
 * ************************** BEGIN LICENSE BLOCK ******************************
 *
 * Copyright © 2023 Christian Stolze
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * ************************** END LICENSE BLOCK ********************************
 */

package webapp

/*

The webapp package provides the communication layer between backend and frontend for the user UI, which will
be served locally as a web app. It is basically a JSON RPC client-server model.

*/

import (

	// standard
	"crypto/tls"
	"fmt"
	"sync"

	// "os"
	// "time"

	// external
	// hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
	"net"
	"net/http"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"

	"github.com/spf13/cobra"

	// internal
	"renderhive/logger"
	// "renderhive/globals"
	// "renderhive/hedera"
)

// structure for the web app manager
type PackageManager struct {

	// JSON RPC
	// General
	Mutex    sync.Mutex
	Listener net.Listener
	Port     string

	// Services
	PingService     *PingService
	OperatorService *OperatorService

	// Placeholder
	Placeholder string

	// Command line interface
	Command      *cobra.Command
	CommandFlags struct {
		FlagPlaceholder bool
	}
}

// WEBAPP MANAGER
// #############################################################################
// create the render manager variable
var Manager = PackageManager{}

// Initialize everything required for the web app management
func (webappm *PackageManager) Init() error {
	var err error

	// log information
	logger.Manager.Package["webapp"].Info().Msg("Initializing the web app manager ...")

	// Create all services
	webappm.OperatorService = new(OperatorService)
	webappm.PingService = new(PingService)

	return err

}

// Deinitialize the web app manager
func (webappm *PackageManager) DeInit() error {
	var err error

	// log event
	logger.Manager.Package["webapp"].Debug().Msg("Deinitializing the web app manager ...")

	return err

}

// Start the JSON RPC server
func (webappm *PackageManager) StartServer(port string, certFile string, keyFile string) error {
	var err error

	// Create the RPC server
	s := rpc.NewServer()

	// set JSON codec
	s.RegisterCodec(json.NewCodec(), "application/json")

	// register all services
	err = s.RegisterService(webappm.PingService, "PingService")
	if err != nil {
		return err
	}
	err = s.RegisterService(webappm.OperatorService, "OperatorService")
	if err != nil {
		return err
	}

	// HTTP handler
	http.HandleFunc("/jsonrpc", func(w http.ResponseWriter, r *http.Request) {

		// log event
		logger.Manager.Package["webapp"].Debug().Msg(fmt.Sprintf("Received request: %s %s", r.Method, r.URL.Path))

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length")

		// Handling CORS preflight request
		if r.Method == "OPTIONS" {
			// log event
			logger.Manager.Package["webapp"].Debug().Msg("Handling the OPTIONS Request")

			w.WriteHeader(http.StatusOK)
			return
		}

		// in case someone opens it in a browser
		if r.Method == http.MethodGet {
			w.Write([]byte("JSON-RPC server active. Please use POST requests for RPC calls."))
			return
		}

		// Handling RPC POST request
		if r.Method == "POST" {
			logger.Manager.Package["webapp"].Debug().Msg("Handling the POST Request")
			s.ServeHTTP(w, r)
			return
		}

	})

	// Setting up HTTPS
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // Ensure modern TLS version
	}
	server := &http.Server{
		Addr:      ":" + port,
		TLSConfig: tlsConfig,
	}

	// log event
	logger.Manager.Package["webapp"].Debug().Msg(fmt.Sprintf("Server starting on port %v ...", webappm.Port))

	// Start the server
	err = server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		return err
	}

	// // start the tcp server
	// webappm.Listener, err = tls.Listen("tcp", ":"+, config)
	// if err != nil {
	// 	return err
	// }
	// defer webappm.StopServer()

	// // store the port
	// webappm.Port = port

	// // listen for incoming connections
	// for {
	// 	conn, err := webappm.Listener.Accept()
	// 	if err != nil {
	// 		logger.Manager.Package["webapp"].Error().Msg(fmt.Sprint("Connection error:", err))
	// 		continue
	// 	}
	// 	go jsonrpc.ServeConn(conn)

	// 	// log event
	// 	logger.Manager.Package["webapp"].Info().Msg(fmt.Sprintf(" [#] Connection between %v and %v established ...", conn.LocalAddr().String(), conn.RemoteAddr().String()))

	// }

	return err

}

// Stop the JSON RPC server
func (webappm *PackageManager) StopServer() {

	// if the server is runnig
	if webappm.Listener != nil {
		webappm.Listener.Close()
	}

	// log event
	logger.Manager.Package["webapp"].Debug().Msg("Server was stopped.")

}

// WEBAPP MANAGER COMMAND LINE INTERFACE
// #############################################################################
// Create the command for the command line interface
func (webappm *PackageManager) CreateCommand() *cobra.Command {

	// create the package command
	webappm.Command = &cobra.Command{
		Use:   "webapp",
		Short: "Commands for the web frontend of the Renderhive Service App",
		Long:  "This command and its sub-commands enable the management of the web frontend for the Renderhive Service App.",
		Run: func(cmd *cobra.Command, args []string) {

			return

		},
	}

	return webappm.Command

}
