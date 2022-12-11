/*
 * ************************** BEGIN LICENSE BLOCK ******************************
 *
 * Copyright © 2022 Christian Stolze
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

package cli

import (
  // "errors"
  "fmt"
  "flag"
  "os"
  "runtime"
)


// MAIN COMMAND LINE INTERFACE
// #############################################################################
// an empty structure to hold all methods
type CommandLine struct {}

func (cli *CommandLine) Help() {
	fmt.Println("Usage:")
	fmt.Println(" node command argument - commands for the renderhive node")
	fmt.Println(" ipfs command argument - commands for the internal IPFS node")
	fmt.Println(" wallet command argument - commands for the internal wallet")
}

func (cli *CommandLine) validate_args() {
	if len(os.Args) < 2 {
		cli.Help()
		runtime.Goexit()
	}
}

func (cli *CommandLine) Start() {

	// nodeID := os.Getenv("NODE_ID")
	// if nodeID == "" {
	// 	fmt.Printf("NODE_ID env is not set!")
	// 	runtime.Goexit()
	// }

  // define main arguments
	helpCmd := flag.Bool("help", false, "Prints the help for the command line interface")
  flag.Parse()

  if *helpCmd {
     cli.Help()
  } else {

    // handle all subcommands
    if err := subcommands(os.Args[1:]); err != nil {
  		fmt.Println(err)
  		os.Exit(1)
  	}
  }
}


// SUBCOMMANDS
// #############################################################################
// an interface for all the subcommands
type Runner interface {
	Init([]string) error
	Run() error
	Name() string
}

func subcommands(args []string) error {
	if len(args) < 1 {
		return nil
	}

  // add all available subcommands to the interface
	cmds := []Runner{
		NewSubCommandNode(),
  	NewSubCommandIpfs(),
  	NewSubCommandWallet(),
	}

  // get the parsed subcommand
	subcommand := os.Args[1]

  // try to find the parsed subcommand among the available subcommands
	for _, cmd := range cmds {
		if cmd.Name() == subcommand {
      // if a subcommand was found, initialize and run it
			cmd.Init(os.Args[2:])
			return cmd.Run()
		}
	}

	return fmt.Errorf("[ERROR] Unknown subcommand: %s", subcommand)
}
