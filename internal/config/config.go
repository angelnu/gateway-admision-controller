package config

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

// CmdConfig represents the configuration of the command.
type CmdConfig struct {
	Debug                bool
	Development          bool
	KeepDNS              bool
	SetGatewayDefault    bool
	WebhookListenAddr    string
	MetricsListenAddr    string
	MetricsPath          string
	TLSCertFilePath      string
	TLSKeyFilePath       string
	Gateway              string
	SetDNSPolicy         string
	SetGatewayLabel      string
	SetGatewayAnnotation string
	InitImage            string
	InitImagePullPol     string
	InitCmd              string
	InitMountPoint       string
	SidecarImage         string
	SidecarImagePullPol  string
	SidecarCmd           string
	SidecarMountPoint    string
	ConfigmapName        string
}

var (
	// Version is set at compile time.
	Version = "dev"
)

// NewCmdConfig returns a new command configuration.
func NewCmdConfig() (*CmdConfig, error) {
	c := &CmdConfig{}
	app := kingpin.New("gateway-admision-controller", "Kubenetes admision controller webhook to change the POD default gateway and DNS")
	app.Version(Version)

	app.Flag("debug", "Enable debug mode.").BoolVar(&c.Debug)
	app.Flag("development", "Enable development mode.").BoolVar(&c.Development)
	app.Flag("webhook-listen-address", "The address where the HTTPS server will be listening to serve the webhooks.").Default(":8080").StringVar(&c.WebhookListenAddr)
	app.Flag("tls-cert-file-path", "The path for the webhook HTTPS server TLS cert file.").StringVar(&c.TLSCertFilePath)
	app.Flag("tls-key-file-path", "The path for the webhook HTTPS server TLS key file.").StringVar(&c.TLSKeyFilePath)

	app.Flag("gateway", "Name/IP of gateway pod").StringVar(&c.Gateway)
	app.Flag("keepDNS", "Do not modify pod DNS").BoolVar(&c.KeepDNS)
	app.Flag("setDNSPolicy", "Set DNSPolicy").StringVar(&c.SetDNSPolicy)
	app.Flag("setGatewayDefault", "Set gateway by default in absence of label/annotation").BoolVar(&c.SetGatewayDefault)
	app.Flag("setGatewayLabel", "Set gateway for pods with this label set to 'true'").StringVar(&c.SetGatewayLabel)
	app.Flag("setGatewayAnnotation", "Set gateway for pods with this annotation set to 'true'").StringVar(&c.SetGatewayAnnotation)

	app.Flag("initImage", "Init container image").StringVar(&c.InitImage)
	app.Flag("initImagePullPol", "Init container pull policy").StringVar(&c.InitImagePullPol)
	app.Flag("initCmd", "Init command to execute instead of container default").StringVar(&c.InitCmd)
	app.Flag("initMountPoint", "Mountpoint for configmap in init container").StringVar(&c.InitMountPoint)

	app.Flag("sidecarImage", "Sidecar container image").StringVar(&c.SidecarImage)
	app.Flag("sidecarImagePullPol", "Sidecar container pull policy").StringVar(&c.SidecarImagePullPol)
	app.Flag("sidecarCmd", "Sidecard command to execute instead of container default").StringVar(&c.SidecarCmd)
	app.Flag("sidecarMountPoint", "Mountpoint for configmap in sidecar container").StringVar(&c.SidecarMountPoint)

	app.Flag("configmapName", "Name of the configmap to attach to containers").StringVar(&c.ConfigmapName)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	return c, nil
}
