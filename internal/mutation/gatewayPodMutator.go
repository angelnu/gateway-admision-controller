package gatewayPodMutator

import (
	"context"
	"net"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k8s-at-home/gateway-admision-controller/internal/resolv"
	kwhmodel "github.com/slok/kubewebhook/v2/pkg/model"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"

	"github.com/k8s-at-home/gateway-admision-controller/internal/config"
	"github.com/k8s-at-home/gateway-admision-controller/internal/log"
)

const (
	GATEWAY_INIT_CONTAINER_NAME    = "gateway-init"
	GATEWAY_SIDECAR_CONTAINER_NAME = "gateway-sidecar"
	GATEWAY_CONFIGMAP_VOLUME_NAME  = "gateway-configmap"
)

var (
	GATEWAY_CONFIGMAP_VOLUME_MODE int32 = 0777
)

type GatewayPodMutator interface {
	GatewayPodMutator(ctx context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error)
}

// NewLabelMarker returns a new marker that will mark with labels.
func NewGatewayPodMutator(cmdConfig config.CmdConfig, logger log.Logger) (GatewayPodMutator, error) {

	logger.Infof("Command config is %#v", cmdConfig)

	if cmdConfig.Gateway != "" {
		//Check we got a valid Gateway
		_, error := net.LookupIP(cmdConfig.Gateway)
		if error != nil {
			return nil, error
		}
	}

	if cmdConfig.DNS != "" {
		//Check we got a valid Gateway
		_, error := net.LookupIP(cmdConfig.DNS)
		if error != nil {
			return nil, error
		}
	}

	DNS_config, error := resolv.Config()
	if error != nil {
		return nil, error
	}
	logger.Infof("Current DNS config is %#v", DNS_config)

	podDNSConfigOptions := make([]corev1.PodDNSConfigOption, 0)
	for i := range DNS_config.Options {
		podDNSConfigOptions = append(podDNSConfigOptions, corev1.PodDNSConfigOption{
			Name:  DNS_config.Options[i].Name,
			Value: DNS_config.Options[i].Value,
		})
	}

	return gatewayPodMutatorCfg{
		cmdConfig: cmdConfig,
		staticDNS: corev1.PodDNSConfig{
			Nameservers: DNS_config.Nameservers,
			Searches:    DNS_config.Search,
			Options:     podDNSConfigOptions,
		},
		logger: logger,
	}, nil
}

func (cfg gatewayPodMutatorCfg) getGatewayIP() (string, error) {
	getGatewayIPs, error := net.LookupIP(cfg.cmdConfig.Gateway)
	return getGatewayIPs[0].String(), error
}

func (cfg gatewayPodMutatorCfg) getDNSIP() (string, error) {
	DNSIPs, error := net.LookupIP(cfg.cmdConfig.DNS)
	return DNSIPs[0].String(), error
}

type gatewayPodMutatorCfg struct {
	cmdConfig config.CmdConfig
	staticDNS corev1.PodDNSConfig
	logger    log.Logger
}

func (cfg gatewayPodMutatorCfg) GatewayPodMutator(_ context.Context, adReview *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {

	pod, ok := obj.(*corev1.Pod)
	if !ok {
		// If not a pod just continue the mutation chain(if there is one) and don't do nothing.
		return &kwhmutating.MutatorResult{}, nil
	}

	setGateway := cfg.cmdConfig.SetGatewayDefault
	var err error

	//Check if label excludes this pod
	if val, ok := pod.GetLabels()[cfg.cmdConfig.SetGatewayLabel]; cfg.cmdConfig.SetGatewayLabel != "" && ok {
		setGateway, err = strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
	}

	//Check if annotations excludes this pod
	if val, ok := pod.GetAnnotations()[cfg.cmdConfig.SetGatewayAnnotation]; cfg.cmdConfig.SetGatewayAnnotation != "" && ok {
		setGateway, err = strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
	}

	if setGateway {

		var error error
		var DNS_IP string
		if cfg.cmdConfig.DNS != "" {
			//Add DNS
			DNS_IP, error = cfg.getDNSIP()
			if error != nil {
				return nil, error
			}

			pod.Spec.DNSConfig = &corev1.PodDNSConfig{
				Nameservers: []string{DNS_IP},
				// Searches: []string{},
				// Options:  []corev1.PodDNSConfigOption{},
			}

			if cfg.cmdConfig.DNSPolicy == "None" {
				// Copy my own webhook settings
				copied := cfg.staticDNS.DeepCopy()

				//fix the first search to match the pod namespace
				for i := range copied.Searches {
					searchParts := strings.Split(copied.Searches[i], ".")
					if len(searchParts) > 2 && searchParts[1] == "svc" {
						if pod.Namespace != "" {
							searchParts[0] = pod.Namespace
							cfg.logger.Infof("corrected namespace in search to POD namespace")
						} else if adReview.Namespace != "" {
							searchParts[0] = adReview.Namespace
							cfg.logger.Infof("corrected namespace in search to adReview namespace")
						} else {
							cfg.logger.Warningf("Empty namespace - not changing search domainss")
						}
						copied.Searches[i] = strings.Join(searchParts, ".")
					}
					if len(copied.Searches[i]) == 0 {
						// circumvention for k3s 1.25
						// https://github.com/angelnu/gateway-admision-controller/issues/54
						copied.Searches[i] = "."
					}
				}

				pod.Spec.DNSConfig.Searches = copied.Searches
				pod.Spec.DNSConfig.Options = copied.Options
			}
		}

		k8s_DNS_ips := strings.Join(cfg.staticDNS.Nameservers, " ")

		if cfg.cmdConfig.DNSPolicy != "" {
			//Add DNSPolicy
			pod.Spec.DNSPolicy = corev1.DNSPolicy(cfg.cmdConfig.DNSPolicy)
		}

		if cfg.cmdConfig.InitImage != "" {

			var volumeMount []corev1.VolumeMount
			if cfg.cmdConfig.InitMountPoint != "" {
				// Create volume mount
				volumeMount = []corev1.VolumeMount{
					corev1.VolumeMount{
						Name:      GATEWAY_CONFIGMAP_VOLUME_NAME,
						ReadOnly:  true,
						MountPath: cfg.cmdConfig.InitMountPoint,
						// SubPath:          "",
						// MountPropagation: &"",
						// SubPathExpr:      "",
					},
				}
			}

			// Create init container
			initContainerRunAsUser := int64(0) // Run init container as root
			initContainerRunAsNonRoot := false
			container := corev1.Container{
				Name:    GATEWAY_INIT_CONTAINER_NAME,
				Image:   cfg.cmdConfig.InitImage,
				Command: []string{cfg.cmdConfig.InitCmd},
				// Args:                     []string{},
				// WorkingDir:               "",
				// Ports:                    []corev1.ContainerPort{},
				// EnvFrom:                  []corev1.EnvFromSource{},
				Env: []corev1.EnvVar{
					{
						Name:  "gateway",
						Value: cfg.cmdConfig.Gateway,
					},
					{
						Name:  "DNS",
						Value: cfg.cmdConfig.DNS,
					},
					{
						Name:  "DNS_ip",
						Value: DNS_IP,
					},
					{
						Name:  "K8S_DNS_ips",
						Value: k8s_DNS_ips,
					},
				},
				// Resources:                corev1.ResourceRequirements{},
				VolumeMounts: volumeMount,
				// VolumeDevices:            []corev1.VolumeDevice{},
				// LivenessProbe:            &corev1.Probe{},
				// ReadinessProbe:           &corev1.Probe{},
				// StartupProbe:             &corev1.Probe{},
				// Lifecycle:                &corev1.Lifecycle{},
				// TerminationMessagePath:   "",
				// TerminationMessagePolicy: "",
				ImagePullPolicy: corev1.PullPolicy(cfg.cmdConfig.InitImagePullPol),
				SecurityContext: &corev1.SecurityContext{
					Capabilities: &corev1.Capabilities{
						Add: []corev1.Capability{
							"NET_ADMIN",
							"NET_RAW",
						},
						Drop: []corev1.Capability{},
					},
					RunAsUser:    &initContainerRunAsUser,
					RunAsNonRoot: &initContainerRunAsNonRoot,
				},
				// Stdin:                    false,
				// StdinOnce:                false,
				// TTY:                      false,
			}

			//Add  initContainer to pod
			pod.Spec.InitContainers = append(pod.Spec.InitContainers, container)
		}

		if cfg.cmdConfig.SidecarImage != "" {

			var volumeMount []corev1.VolumeMount
			if cfg.cmdConfig.SidecarMountPoint != "" {
				// Create volume mount
				volumeMount = []corev1.VolumeMount{
					corev1.VolumeMount{
						Name:      GATEWAY_CONFIGMAP_VOLUME_NAME,
						ReadOnly:  true,
						MountPath: cfg.cmdConfig.SidecarMountPoint,
						// SubPath:          "",
						// MountPropagation: &"",
						// SubPathExpr:      "",
					},
				}
			}

			// Create sidecar container
			var sidecarContainerRunAsUser = int64(0) // Run init container as root
			var sidecarContainerRunAsNonRoot = false
			container := corev1.Container{
				Name:    GATEWAY_SIDECAR_CONTAINER_NAME,
				Image:   cfg.cmdConfig.SidecarImage,
				Command: []string{cfg.cmdConfig.SidecarCmd},
				// Args:                     []string{},
				// WorkingDir:               "",
				// Ports:                    []corev1.ContainerPort{},
				// EnvFrom:                  []corev1.EnvFromSource{},
				Env: []corev1.EnvVar{
					{
						Name:  "gateway",
						Value: cfg.cmdConfig.Gateway,
					},
					{
						Name:  "DNS",
						Value: cfg.cmdConfig.DNS,
					},
					{
						Name:  "DNS_ip",
						Value: DNS_IP,
					},
					{
						Name:  "K8S_DNS_ips",
						Value: k8s_DNS_ips,
					},
				},
				// Resources:                corev1.ResourceRequirements{},
				VolumeMounts: volumeMount,
				// VolumeDevices:            []corev1.VolumeDevice{},
				// LivenessProbe:            &corev1.Probe{},
				// ReadinessProbe:           &corev1.Probe{},
				// StartupProbe:             &corev1.Probe{},
				// Lifecycle:                &corev1.Lifecycle{},
				// TerminationMessagePath:   "",
				// TerminationMessagePolicy: "",
				ImagePullPolicy: corev1.PullPolicy(cfg.cmdConfig.SidecarImagePullPol),
				SecurityContext: &corev1.SecurityContext{
					Capabilities: &corev1.Capabilities{
						Add: []corev1.Capability{
							"NET_ADMIN",
							"NET_RAW",
						},
						Drop: []corev1.Capability{},
					},
					RunAsUser:    &sidecarContainerRunAsUser,
					RunAsNonRoot: &sidecarContainerRunAsNonRoot,
				},
				// Stdin:                    false,
				// StdinOnce:                false,
				// TTY:                      false,
			}

			//Add container to pod
			pod.Spec.Containers = append(pod.Spec.Containers, container)
		}

		if cfg.cmdConfig.ConfigmapName != "" {
			pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
				Name: GATEWAY_CONFIGMAP_VOLUME_NAME,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cfg.cmdConfig.ConfigmapName,
						},
						DefaultMode: &GATEWAY_CONFIGMAP_VOLUME_MODE,
					},
				},
			})
		}
	}

	cfg.logger.Infof("Mutated pod %s", pod.Name)
	cfg.logger.Debugf("%s", pod.String())

	return &kwhmutating.MutatorResult{
		MutatedObject: pod,
	}, nil

}
