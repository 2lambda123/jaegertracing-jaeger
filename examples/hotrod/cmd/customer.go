// Copyright © 2017 Uber Technologies, Inc.
//

package cmd

import (
	"net"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/uber-go/zap"

	"code.uber.internal/infra/jaeger-demo/pkg/log"
	"code.uber.internal/infra/jaeger-demo/pkg/tracing"
	"code.uber.internal/infra/jaeger-demo/services/customer"
)

// customerCmd represents the customer command
var customerCmd = &cobra.Command{
	Use:   "customer",
	Short: "Starts Customer service",
	Long:  `Starts Customer service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.NewFactory(logger.With(zap.String("service", "customer")))
		server := customer.NewServer(
			net.JoinHostPort(customerOptions.serverInterface, strconv.Itoa(customerOptions.serverPort)),
			tracing.Init("customer", logger),
			logger,
		)
		return server.Run()
	},
}

var (
	customerOptions struct {
		serverInterface string
		serverPort      int
	}
)

func init() {
	RootCmd.AddCommand(customerCmd)

	customerCmd.Flags().StringVarP(&customerOptions.serverInterface, "bind", "", "127.0.0.1", "interface to which the Customer server will bind")
	customerCmd.Flags().IntVarP(&customerOptions.serverPort, "port", "p", 8081, "port on which the Customer server will listen")
}
