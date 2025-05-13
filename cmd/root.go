package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/syedomair/istio-config-manager/config"
	"github.com/syedomair/istio-config-manager/istio"
	"github.com/syedomair/istio-config-manager/kube"
)

var action string

func newRootCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "istio-config-manager",
		Short: "Automates L4/L7 Istio configuration",
		Run:   runRootCmd,
	}
	cmd.Flags().StringVarP(&action, "action", "a", "", "Istio config action to apply (h_o, t_s, d_f, t_m, t_o, r_p, c_b)")
	cmd.MarkFlagRequired("action")
	return cmd

}

func runRootCmd(cmd *cobra.Command, args []string) {
	if err := applyIstioConfig(action); err != nil {
		fmt.Println("Error applying config:", err)
		os.Exit(1)
	}
	fmt.Println("Istio configuration applied successfully.")
}

func applyIstioConfig(action string) error {
	conf := config.LoadConfig()
	clients := kube.MustClients()

	var err error

	switch action {
	case "h_o":
		err = istio.ApplyHeaderOverride(clients.Istio, conf)
	case "t_s":
		err = istio.ApplyTrafficSplit(clients.Istio, conf)
	case "d_f":
		err = istio.ApplyDelayFault(clients.Istio, conf)
	case "t_m":
		err = istio.CreateTrafficMirroring(clients.Istio, conf)
	case "t_o":
		err = istio.ApplyTimeout(clients.Istio, conf)
	case "r_p":
		err = istio.ApplyRetryPolicy(clients.Istio, conf)
	case "c_b":
		err = istio.ApplyCircuitBreaker(clients.Istio, conf)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	if err != nil {
		return fmt.Errorf("failed to apply Istio configuration for action %s: %w", action, err)
	}
	return nil
}

func Execute() {
	cobra.CheckErr(newRootCmd().Execute())
}
