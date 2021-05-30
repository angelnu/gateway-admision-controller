package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

// CmdConfig represents the configuration of the command.
type CmdConfig struct {
	Debug             bool
	Development       bool
	WebhookListenAddr string
	MetricsListenAddr string
	MetricsPath       string
	TLSCertFilePath   string
	TLSKeyFilePath    string
	gateway           string
	keepDNS           bool
}

// NewCmdConfig returns a new command configuration.
func NewCmdConfig() (*CmdConfig, error) {
	c := &CmdConfig{}
	app := kingpin.New("gateway-admision-controller", "Kubenetes admision controller webhook to change the POD default gateway and DNS")
	app.Version(Version)

	app.Flag("debug", "Enable debug mode.").BoolVar(&c.Debug)
	app.Flag("development", "Enable development mode.").BoolVar(&c.Development)
	app.Flag("webhook-listen-address", "the address where the HTTPS server will be listening to serve the webhooks.").Default(":8080").StringVar(&c.WebhookListenAddr)
	app.Flag("tls-cert-file-path", "the path for the webhook HTTPS server TLS cert file.").StringVar(&c.TLSCertFilePath)
	app.Flag("tls-key-file-path", "the path for the webhook HTTPS server TLS key file.").StringVar(&c.TLSKeyFilePath)
	app.Flag("gateway", "name/IP of gateway pod").StringVar(&c.gateway)
	app.Flag("keepDNS", "do not modify pod DNS").BoolVar(&c.keepDNS)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	return c, nil
}
