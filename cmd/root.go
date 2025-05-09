package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/syedomair/istio-config-manager/config"
	"github.com/syedomair/istio-config-manager/istio"
	"github.com/syedomair/istio-config-manager/kube"
)

func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "istio-config-manager",
		Short: "Automates L4/L7 Istio configuration",
		Run:   runRootCmd,
	}
}

func runRootCmd(cmd *cobra.Command, args []string) {
	if err := applyIstioConfig(); err != nil {
		fmt.Println("Error applying config:", err)
		os.Exit(1)
	}
	fmt.Println("Istio configuration applied successfully.")
}

func applyIstioConfig() error {
	conf := config.LoadConfig()
	clients := kube.MustClients()
	//return istio.ApplyConfig(clients, conf)
	return istio.ApplyHeaderOverride(clients.Istio, conf)
}

func Execute() {
	cobra.CheckErr(newRootCmd().Execute())
}
