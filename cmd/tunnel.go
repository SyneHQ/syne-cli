package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"syne-cli/internal/tunnel"

	"github.com/spf13/cobra"
)

var (
	sshHost     string
	sshPort     string
	sshUser     string
	sshKeyPath  string
	localPort   string
	remoteHost  string
	remotePort  string
	bindAddress string
)

func init() {
	tunnelCmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Create an SSH tunnel for PostgreSQL connections",
		Long: `Create a secure SSH tunnel to allow cloud server to connect to your local PostgreSQL instance.
This command will establish a secure tunnel between your local machine and our cloud server.
The cloud server will be able to connect to your local PostgreSQL instance through this encrypted tunnel.

Example:
  syne-cli tunnel \
    --ssh-host cloud.example.com \
    --ssh-user tunnel \
    --remote-port 5432`,
		RunE: runTunnel,
	}

	// SSH connection flags
	tunnelCmd.Flags().StringVar(&sshHost, "ssh-host", "", "SSH server hostname (required)")
	tunnelCmd.Flags().StringVar(&sshPort, "ssh-port", "22", "SSH server port")
	tunnelCmd.Flags().StringVar(&sshUser, "ssh-user", "", "SSH username (required)")
	tunnelCmd.Flags().StringVar(&sshKeyPath, "ssh-key", "", "Path to SSH private key (defaults to ~/.ssh/id_rsa)")

	// Port forwarding flags
	tunnelCmd.Flags().StringVar(&localPort, "local-port", "5432", "Local PostgreSQL port to forward")
	tunnelCmd.Flags().StringVar(&remoteHost, "remote-host", "localhost", "Remote host for the tunnel endpoint")
	tunnelCmd.Flags().StringVar(&remotePort, "remote-port", "", "Remote port on the cloud server (required)")
	tunnelCmd.Flags().StringVar(&bindAddress, "bind", "localhost", "Local address to bind to")

	// Required flags
	tunnelCmd.MarkFlagRequired("ssh-host")
	tunnelCmd.MarkFlagRequired("ssh-user")
	tunnelCmd.MarkFlagRequired("remote-port")

	rootCmd.AddCommand(tunnelCmd)
}

func runTunnel(cmd *cobra.Command, args []string) error {
	// Create tunnel configuration
	config := tunnel.TunnelConfig{
		SSHHost:     sshHost,
		SSHPort:     sshPort,
		SSHUser:     sshUser,
		SSHKeyPath:  sshKeyPath,
		LocalPort:   localPort,
		RemoteHost:  remoteHost,
		RemotePort:  remotePort,
		BindAddress: bindAddress,
	}

	// Create and start tunnel
	t, err := tunnel.NewTunnel(config)
	if err != nil {
		return fmt.Errorf("error creating tunnel: %v", err)
	}

	if err := t.Start(); err != nil {
		return fmt.Errorf("error starting tunnel: %v", err)
	}

	// Handle interrupt signal for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("\nTunnel is running. Press Ctrl+C to stop.\n")

	// Wait for interrupt signal
	<-sigChan

	fmt.Printf("\nStopping tunnel...\n")
	return t.Stop()
}
