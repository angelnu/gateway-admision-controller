package gatewayPodMutator

import (
	"context"
	"fmt"
	"net"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kwhmodel "github.com/slok/kubewebhook/v2/pkg/model"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"

	"github.com/k8s-at-home/gateway-admision-controller/internal/config"
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
func NewGatewayPodMutator(cmdConfig config.CmdConfig) (GatewayPodMutator, error) {

	if cmdConfig.Gateway == "" {
		return nil, fmt.Errorf("gateway is required")
	}

	gatewayIPs, error := net.LookupIP(cmdConfig.Gateway)
	if error != nil {
		return nil, error
	}

	return gatewayPodMutatorCfg{
		cmdConfig:  cmdConfig,
		gatewayIPs: gatewayIPs,
	}, nil
}

func (cfg gatewayPodMutatorCfg) getGatewayIP() (string, error) {
	gatewayIPs, error := net.LookupIP(cfg.cmdConfig.Gateway)
	if error != nil {
		return "", error
	}
	return gatewayIPs[0].String(), nil
}

type gatewayPodMutatorCfg struct {
	cmdConfig  config.CmdConfig
	gatewayIPs []net.IP
}

func (cfg gatewayPodMutatorCfg) GatewayPodMutator(_ context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {
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
					Privileged: &[]bool{true}[0],
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

			// Create init container
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
					Privileged: &[]bool{true}[0],
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

		if !cfg.cmdConfig.KeepDNS {

			gatewayIP, error := cfg.getGatewayIP()
			if error != nil {
				return nil, error
			}
			//Add DNS
			pod.Spec.DNSPolicy = corev1.DNSPolicy(cfg.cmdConfig.SetDNSPolicy)
			pod.Spec.DNSConfig = &corev1.PodDNSConfig{
				Nameservers: []string{gatewayIP},
				// Searches: []string{},
				// Options:  []corev1.PodDNSConfigOption{},
			}
		}
	}

	return &kwhmutating.MutatorResult{
		MutatedObject: pod,
	}, nil

}
