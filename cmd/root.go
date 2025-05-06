package cmd

import (
	"fmt"
	"os"

	"github.com/syedomair/istio-config-manager/config"
	"github.com/syedomair/istio-config-manager/istio"
	"github.com/syedomair/istio-config-manager/kube"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "istio-config-manager",
	Short: "Automates L4/L7 Istio configuration",
	Run: func(cmd *cobra.Command, args []string) {
		conf := config.LoadConfig()
		client := kube.MustClient()
		if err := istio.ApplyConfig(client, conf); err != nil {
			fmt.Println("Error applying config:", err)
			os.Exit(1)
		}
		fmt.Println("Istio configuration applied successfully.")
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
