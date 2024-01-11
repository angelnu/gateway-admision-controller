package config

import (
	"os"

	"github.com/alecthomas/kingpin/v2"
)

// CmdConfig represents the configuration of the command.
type CmdConfig struct {
	Debug                     bool
	Development               bool
	SetGatewayDefault         bool
	WebhookListenAddr         string
	MetricsListenAddr         string
	MetricsPath               string
	TLSCertFilePath           string
	TLSKeyFilePath            string
	Gateway                   string
	DNS                       string
	DNSPolicy                 string
	SetGatewayLabel           string
	SetGatewayLabelValue      string
	SetGatewayAnnotation      string
	SetGatewayAnnotationValue string
	InitImage                 string
	InitImagePullPol          string
	InitImagePrepend          bool
	InitCmd                   string
	InitMountPoint            string
	SidecarImage              string
	SidecarImagePullPol       string
	SidecarCmd                string
	SidecarMountPoint         string
	ConfigmapName             string
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

	app.Flag("gateway", "Name/IP of the gateway pod").StringVar(&c.Gateway)
	app.Flag("DNS", "Name/IP of the DNS (might be the same as the gateway pod)").StringVar(&c.DNS)
	app.Flag("DNSPolicy", "Set DNSPolicy").StringVar(&c.DNSPolicy)

	app.Flag("setGatewayDefault", "Set gateway by default in absence of label/annotation").BoolVar(&c.SetGatewayDefault)
	app.Flag("setGatewayLabel", "Set gateway for pods with this label set to 'true'").StringVar(&c.SetGatewayLabel)
	app.Flag("setGatewayLabelValue", "Set gateway for pods with label set to this value").StringVar(&c.SetGatewayLabelValue)
	app.Flag("setGatewayAnnotation", "Set gateway for pods with this annotation set to 'true'").StringVar(&c.SetGatewayAnnotation)
	app.Flag("setGatewayAnnotationValue", "Set gateway for pods with annotation set to this value").StringVar(&c.SetGatewayAnnotationValue)

	app.Flag("initImage", "Init container image").StringVar(&c.InitImage)
	app.Flag("initImagePullPol", "Init container pull policy").StringVar(&c.InitImagePullPol)
	app.Flag("initCmd", "Init command to execute instead of container default").StringVar(&c.InitCmd)
	app.Flag("initMountPoint", "Mountpoint for configmap in init container").StringVar(&c.InitMountPoint)
	app.Flag("initImagePrepend", "Prepend or append to container").Default().BoolVar(&c.InitImagePrepend)

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
