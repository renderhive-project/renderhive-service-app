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
  "renderhive/node"
  "renderhive/hedera"
  "renderhive/ipfs"
  "renderhive/renderer"
  "renderhive/webapp"

)



// CLI STRUCTURES, VARIABLES & CONSTANTS
// #############################################################################
type Commands struct {

  // renderhive main commands
  Main *cobra.Command
  MainFlags struct {

    Interactive bool

  }

  // subcommands
  Help *cobra.Command
  Exit *cobra.Command

}

type PackageManager struct {

  // CLI commands
  Commands Commands
  Quit bool

}


// CLI MANAGER
// #############################################################################
// create the render manager variable
var Manager = PackageManager{}

// Initialize the CLI Manager
func (clim *PackageManager) Init() (error) {
    var err error

    logger.Manager.Package["cli"].Debug().Msg("Initializing the Command Line Interface manager ...")

    // create the main command
    clim.Commands.Main = clim.CreateMainCommand()

    // for each package, add the package command to the CLI
    clim.AddPackageCommand(node.Manager.CreateCommand())
    clim.AddPackageCommand(hedera.Manager.CreateCommand())
    clim.AddPackageCommand(ipfs.Manager.CreateCommand())
    clim.AddPackageCommand(renderer.Manager.CreateCommand())
    clim.AddPackageCommand(webapp.Manager.CreateCommand())

    return err
}

// Deinitialize the CLI manager
func (clim *PackageManager) DeInit() (error) {
    var err error

    // log event
    logger.Manager.Package["cli"].Debug().Msg("Deinitializing the Command Line Interface manager ...")

    return err

}


// COMMAND LINE INTERFACE
// #############################################################################
// Create the main command for the command line interface
func (clim *PackageManager) CreateMainCommand() *cobra.Command {

    // log debug event
    logger.Manager.Package["cli"].Debug().Msg("Create the main commands for the CLI.")

    // create the main command
    clim.Commands.Main = &cobra.Command{
      Use:     "renderhive",
      Short:   "Renderhive is a crowdrendering plattform for Blender based on Web3 technologies",
      Long:    "This command line interface gives you complete control over the Renderhive Service App backend, which is the main software package for participating in the Renderhive network – the first crowdrendering platform for Blender built on Web3 technologies.",
      RunE: func(cmd *cobra.Command, args []string) error {

        // if the app was started in interactive mode
        if (clim.Commands.MainFlags.Interactive) {

          // new command line
      		fmt.Println("------------------------------------------------------------")
      		fmt.Println("|    _____                _           _     _              |")
      		fmt.Println("|   |  __ \\              | |         | |   (_)             |")
      		fmt.Println("|   | |__) |___ _ __   __| | ___ _ __| |__  ___   _____    |")
      		fmt.Println("|   |  _  // _ \\ '_ \\ / _` |/ _ \\ '__| '_ \\| \\ \\ / / _ \\   |")
      		fmt.Println("|   | | \\ \\  __/ | | | (_| |  __/ |  | | | | |\\ V /  __/   |")
      		fmt.Println("|   |_|  \\_\\___|_| |_|\\__,_|\\___|_|  |_| |_|_| \\_/ \\___|   |")
      		fmt.Println("|                  COMMAND LINE INTERFACE                  |")
      		fmt.Println("------------------------------------------------------------")
      		fmt.Println("")
      		fmt.Println("Interact with the Renderhive network from the command line:")

          // start a session for user interaction
      		for !clim.Quit {

            // new command line
      			fmt.Print("(renderhive) > ")

            // wait for new user input
      			reader := bufio.NewReader(os.Stdin)
      			input, _ := reader.ReadString('\n')

            // split the input into arguments
      			input = strings.TrimSpace(input)
      			args := strings.Split(input, " ")

            // process the command
            clim.ProcessSessionCommand(clim.Commands.Main, args)

            // empty args again, so that they don't interfere with the next loop
            args = []string{}

      		}

        }

        return nil

      },
    }

    // add command flags
    clim.Commands.Main.PersistentFlags().BoolVarP(&clim.Commands.MainFlags.Interactive, "interactive", "i", false, "Run the Renderhive Service App in an interactive session")

    // Create an 'exit' command for the CLI session
    clim.Commands.Exit = &cobra.Command{
    	Use:   "exit",
    	Short: "Exit the Renderhive command line interface session",
    	Long: "This command will close the command line interface session and shutdown the Renderhive Service App",
    	Run: func(cmd *cobra.Command, args []string) {

        // quit the session
        clim.Quit = true

        return
    	},
    }

    // Create an 'help' command for the CLI session
    clim.Commands.Help = &cobra.Command{
    	Use:   "help",
    	Short: "Print the help for this command line interface",
    	// Long: "This command will close the command line interface session and shutdown the Renderhive Service App",
    	Run: func(cmd *cobra.Command, args []string) {

        // execute the main command with the "help" flag
        fmt.Println("")
        clim.Commands.Main.Help()
        fmt.Println("")

        return
    	},
    }

    // Parse the flags passed to the CLI
    clim.Commands.Main.ParseFlags(os.Args[1:])

    // if the app was NOT started in interactive mode
    if (!clim.Commands.MainFlags.Interactive) {

        // add the command
        clim.Commands.Main.AddCommand(clim.Commands.Exit)
        clim.Commands.Main.AddCommand(clim.Commands.Help)

    }

    return clim.Commands.Main

}

// Create the package command for the command line interface
func (clim *PackageManager) AddPackageCommand(command *cobra.Command) *cobra.Command {
    var packageCommands *cobra.Command

    // log debug event
    logger.Manager.Package["cli"].Debug().Msg("Create the package commands for the CLI.")

    // if the app was NOT started in interactive mode
    if (!clim.Commands.MainFlags.Interactive) {

        // add the commands to the main command
        clim.Commands.Main.AddCommand(command)

    }


    return packageCommands

}

// Create the package command for the command line interface
func (clim *PackageManager) ProcessSessionCommand(cmd *cobra.Command, args []string) {


    // 'exit' command
    if strings.EqualFold(args[0], "exit") {
      clim.Commands.Exit.SetArgs(args)
      clim.Commands.Exit.Execute()

    // 'help' command
    } else if strings.EqualFold(args[0], "help") {
      clim.Commands.Help.SetArgs(args)
      clim.Commands.Help.Execute()

    // Unknown command
    } else {

      // Print an error
      fmt.Println("")
      fmt.Println("Error: Unknown command.")
      fmt.Println("")

    }

}
