package gatewayPodMutator

import (
	"context"
	"net"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kwhmodel "github.com/slok/kubewebhook/v2/pkg/model"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
)

type GatewayPodMutator interface {
	GatewayPodMutator(ctx context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error)
}

// NewLabelMarker returns a new marker that will mark with labels.
func NewGatewayPodMutator(
	gateway string,
	setGatewayLabel string,
	setGatewayAnnotation string,
	keepDNS bool,
	setGatewayDefault bool,
) (GatewayPodMutator, error) {
	gatewayIPs, error := net.LookupIP(gateway)
	if error != nil {
		return nil, error
	}
	return gatewayPodMutatorCfg{
		gatewayIPs:           gatewayIPs,
		setGatewayLabel:      setGatewayLabel,
		setGatewayAnnotation: setGatewayAnnotation,
		keepDNS:              keepDNS,
		setGatewayDefault:    setGatewayDefault,
	}, nil
}

type gatewayPodMutatorCfg struct {
	gatewayIPs           []net.IP
	setGatewayLabel      string
	setGatewayAnnotation string
	keepDNS              bool
	setGatewayDefault    bool
}

func (cfg gatewayPodMutatorCfg) GatewayPodMutator(_ context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		// If not a pod just continue the mutation chain(if there is one) and don't do nothing.
		return &kwhmutating.MutatorResult{}, nil
	}

	setGateway := cfg.setGatewayDefault
	var err error

	//Check if label excludes this pod
	if val, ok := pod.GetLabels()[cfg.setGatewayLabel]; cfg.setGatewayLabel != "" && ok {
		setGateway, err = strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
	}

	//Check if annotations excludes this pod
	if val, ok := pod.GetAnnotations()[cfg.setGatewayAnnotation]; cfg.setGatewayAnnotation != "" && ok {
		setGateway, err = strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
	}

	if setGateway {
		// Create init container
		container := corev1.Container{
			Name:    "add-gateway",
			Image:   "alpine",
			Command: []string{"ip"},
			Args: append(
				strings.Split("route change default via", " "),
				cfg.gatewayIPs[0].String(),
			),
			// WorkingDir:               "",
			// Ports:                    []corev1.ContainerPort{},
			// EnvFrom:                  []corev1.EnvFromSource{},
			// Env:                      []corev1.EnvVar{},
			// Resources:                corev1.ResourceRequirements{},
			// VolumeMounts:             []corev1.VolumeMount{},
			// VolumeDevices:            []corev1.VolumeDevice{},
			// LivenessProbe:            &corev1.Probe{},
			// ReadinessProbe:           &corev1.Probe{},
			// StartupProbe:             &corev1.Probe{},
			// Lifecycle:                &corev1.Lifecycle{},
			// TerminationMessagePath:   "",
			// TerminationMessagePolicy: "",
			// ImagePullPolicy:          "",
			SecurityContext: &corev1.SecurityContext{
				Privileged: &[]bool{true}[0],
			},
			// Stdin:                    false,
			// StdinOnce:                false,
			// TTY:                      false,
		}

		//Add initContainer to pod
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, container)

		if !cfg.keepDNS {
			//Add DNS
			pod.Spec.DNSPolicy = "None"
			pod.Spec.DNSConfig = &corev1.PodDNSConfig{
				Nameservers: []string{
					cfg.gatewayIPs[0].String(),
				},
				// Searches: []string{},
				// Options:  []corev1.PodDNSConfigOption{},
			}
		}
	}

	return &kwhmutating.MutatorResult{
		MutatedObject: pod,
	}, nil

}
