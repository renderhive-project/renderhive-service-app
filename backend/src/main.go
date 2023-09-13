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

package main

import (

	// standard
	"fmt"
	"net/http"
	"os"
	"sync"

	// external
	// hederasdk "github.com/hashgraph/hedera-sdk-go/v2"

	// internal
	// . "renderhive/globals"
	"renderhive/logger"
	// "renderhive/cli"
	// "renderhive/hedera"
	// "renderhive/node"
)

// error value
var err error
var ServiceApp AppManager

// INITIALIZE APP
// #############################################################################
func init() {

	// INITIALIZE SERVICE APP
	// ***************************************************************************
	// TODO: use the signal library to catch interrupts, so that the app still
	//       shuts down decently?
	ServiceApp = AppManager{}
	ServiceApp.Quit = make(chan bool, 1)
	ServiceApp.WG = sync.WaitGroup{}

	// initialize service app
	err = ServiceApp.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

// MAIN FUNCTION
// #############################################################################
func main() {

	// prepare end of program
	defer os.Exit(0)

	// deinitialize the service app at the end of the main function
	defer ServiceApp.DeInit()

	// placeholder
	// fmt.Println(time.Now().Add(30 * time.Second))

	// COMMAND LINE INTERFACE
	// ***************************************************************************
	// if the app was started in interactive CLI mode
	if ServiceApp.CLIManager.Commands.MainFlags.Interactive {

		// start the command line interface
		ServiceApp.CLIManager.StartInteractive()

	}

	// BACKEND SERVER
	// ***************************************************************************
	err := ServiceApp.WebAppManager.StartServer("5174", "webapp/cert/cert.pem", "webapp/cert/key.pem")
	if err != nil {
		if err == http.ErrServerClosed {

			// log information
			logger.Manager.Package["webapp"].Error().Msg("Server closed gracefully")

		} else {

			// log information
			logger.Manager.Package["webapp"].Error().Msg(fmt.Sprintf("Error starting server:", err))

		}

	}

}
