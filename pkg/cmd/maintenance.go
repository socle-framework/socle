package cmd

import (
	"fmt"
	"net/rpc"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(EnableMaintenanceModeCmd)
	rootCmd.AddCommand(DisableMaintenanceModeCmd)
}

var EnableMaintenanceModeCmd = &cobra.Command{
	Use:   "down",
	Short: "Enable maintenance mode",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		inMaintenanceMode(true)
	},
}

var DisableMaintenanceModeCmd = &cobra.Command{
	Use:   "up",
	Short: "Disable maintenance mode",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		inMaintenanceMode(false)
	},
}

func inMaintenanceMode(mode bool) {
	rpcPort := os.Getenv("RPC_PORT")
	c, err := rpc.Dial("tcp", "127.0.0.1:"+rpcPort)
	if err != nil {
		exitGracefully(err)
	}

	fmt.Println("Connected...")
	var result string
	err = c.Call("RPCServer.MaintenanceMode", mode, &result)
	if err != nil {
		exitGracefully(err)
	}

	color.Yellow(result)
}
