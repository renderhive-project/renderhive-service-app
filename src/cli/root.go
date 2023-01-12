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

package cli

/*

  This package contains the command-line interface of the Renderhive service app.
  It is mainly usefull for the development phase, where no GUI is available, but
  it will also enable to run the service app in headless mode later.

*/

import (

  // standard
  "fmt"
  "os"
  "bufio"
  "strings"

  // external
  // hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
  "github.com/spf13/cobra"

  // internal
  // . "renderhive/globals"
  "renderhive/logger"

)



// CLI STRUCTURES, VARIABLES & CONSTANTS
// #############################################################################
type Commands struct {

  // renderhive main commands
  Main *cobra.Command
  MainFlags struct {

    Background bool

  }

  // renderhive bash commands
  // NOTE: The bash will give a command line interface for communicating with
  //       the Renderhive Service App when running it in background / headless
  //       mode.
  bash struct {

    // renderhive package commands
    // NOTE: Each package command will have several subcommands and flags.
    Hedera *cobra.Command
    Ipfs *cobra.Command
    Node *cobra.Command
    Renderer *cobra.Command

    // misc
    Exit *cobra.Command

  }

}

type CLIManager struct {

  // CLI commands
  Commands Commands

}


// CLI MANAGER
// #############################################################################
// Initialize the CLI Manager
func (clim *CLIManager) Init() (error) {
    var err error

    logger.Manager.Package["cli"].Debug().Msg("Initializing the Command Line Interface manager ...")

    // create the main command
    clim.Commands.Main = clim.CreateMainCommand()

    // TODO: create package commands
    clim.CreatePackageCommands()

    // TODO: add package commands
    // ...

    return err
}

// Deinitialize the CLI manager
func (clim *CLIManager) DeInit() (error) {
    var err error

    // log event
    logger.Manager.Package["cli"].Debug().Msg("Deinitializing the Command Line Interface manager ...")

    return err

}


// COMMAND LINE INTERFACE
// #############################################################################
// Create the main command for the command line interface
func (clim *CLIManager) CreateMainCommand() *cobra.Command {

    // log debug event
    logger.Manager.Package["cli"].Debug().Msg("Create the main commands for the CLI.")

    // create the main command
    clim.Commands.Main = &cobra.Command{
      Use:     `renderhive`,
      Short:   `Renderhive is a crowdrendering plattform for Blender based on Web3 technologies`,
      Long:    `This command line interface gives a complete control over the Renderhive Service App backend, which is the main software package for participating in the renderhive.`,
      // Args:    ArgsValidator(config),
      // PreRunE: OptionsValidator(config, headers),
      RunE: func(cmd *cobra.Command, args []string) error {

        // if the app was started in background mode
        if clim.Commands.MainFlags.Background {

          // new command line
    			fmt.Println("------------------------------------------------------------")
    			fmt.Println("|    _____                _           _     _              |")
    			fmt.Println("|   |  __ \\              | |         | |   (_)             |")
    			fmt.Println("|   | |__) |___ _ __   __| | ___ _ __| |__  ___   _____    |")
    			fmt.Println("|   |  _  // _ \\ '_ \\ / _` |/ _ \\ '__| '_ \\| \\ \\ / / _ \\   |")
    			fmt.Println("|   | | \\ \\  __/ | | | (_| |  __/ |  | | | | |\\ V /  __/   |")
    			fmt.Println("|   |_|  \\_\\___|_| |_|\\__,_|\\___|_|  |_| |_|_| \\_/ \\___|   |")
    			fmt.Println("|           SERVICE APP COMMAND LINE INTERFACE             |")
    			fmt.Println("------------------------------------------------------------")
    			fmt.Println("")
    			fmt.Println("Interact with the Renderhive network from the command line:")

          // start a bash for user interaction
      		shouldExit := false
      		for !shouldExit {

            // new command line
      			fmt.Print("(renderhive) > ")

            // wait for new user input
      			reader := bufio.NewReader(os.Stdin)
      			input, _ := reader.ReadString('\n')

            // split the input into arguments
      			input = strings.TrimSpace(input)
      			args := strings.Split(input, " ")

            // set arguments and execute the command
            fmt.Println(args)

      			// rootCmd.SetArgs(args)
      			// rootCmd.Execute()
      		}
      	}

        return nil
      },
    }

    // add command flags
    clim.Commands.Main.PersistentFlags().BoolVarP(&clim.Commands.MainFlags.Background, "background", "b", false, "Run the Renderhive Service App in background / headless mode")

    return clim.Commands.Main

}

// Create the package command for the command line interface
func (clim *CLIManager) CreatePackageCommands() *cobra.Command {
    var packageCommands *cobra.Command


    // log debug event
    logger.Manager.Package["cli"].Debug().Msg("Create the package commands for the CLI.")

    return packageCommands

}


// package main
//
// import (
// 	"bufio"
// 	"fmt"
// 	"os"
// 	"strings"
//
// 	"github.com/spf13/cobra"
// )
//
// var shouldExit bool
//
// var cmdExit = &cobra.Command{
// 	Use:   "exit",
// 	Short: "Exit the input loop",
// 	Long: `Exit the input loop`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		shouldExit = true
// 	},
// }
//
// var cmdInput = &cobra.Command{
// 	Use:   "input",
// 	Short: "Prompt for user input",
// 	Long: `Prompt for user input and execute it as a command.
// This command prompts for user input and executes it as a command.`,
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		shouldExit = false
// 		for !shouldExit {
// 			fmt.Print("> ")
// 			reader := bufio.NewReader(os.Stdin)
// 			input, _ := reader.ReadString('\n')
// 			input = strings.TrimSpace(input)
// 			args := strings.Split(input, " ")
// 			rootCmd.SetArgs(args)
// 			rootCmd.Execute()
// 		}
// 		return nil
// 	},
// }
//
// var rootCmd = &cobra.Command{Use: "app"}
// rootCmd.AddCommand(cmdInput)
// rootCmd.AddCommand(cmdExit)
//
// func main() {
// 	rootCmd.Execute()
// }
