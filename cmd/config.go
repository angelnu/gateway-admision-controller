package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

// CmdConfig represents the configuration of the command.
type CmdConfig struct {
	Debug                 bool
	Development           bool
	keepDNS               bool
	WebhookListenAddr     string
	MetricsListenAddr     string
	MetricsPath           string
	TLSCertFilePath       string
	TLSKeyFilePath        string
	gateway               string
	keepGatewayLabel      string
	keepGatewayAnnotation string
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
	app.Flag("keepGatewayLabel", "do not mutate pods with this label set to 'true'").StringVar(&c.keepGatewayLabel)
	app.Flag("keepGatewayAnnotation", "do not mutate pods with this annotation set to 'true'").StringVar(&c.keepGatewayAnnotation)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	return c, nil
}
