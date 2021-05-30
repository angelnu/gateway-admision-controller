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
	keepGatewayLabel string,
	keepGatewayAnnotation string,
	keepDNS bool,
) (GatewayPodMutator, error) {
	gatewayIPs, error := net.LookupIP(gateway)
	if error != nil {
		return nil, error
	}
	return gatewayPodMutatorCfg{
		gatewayIPs:            gatewayIPs,
		keepGatewayLabel:      keepGatewayLabel,
		keepGatewayAnnotation: keepGatewayAnnotation,
		keepDNS:               keepDNS,
	}, nil
}

type gatewayPodMutatorCfg struct {
	gatewayIPs            []net.IP
	keepGatewayLabel      string
	keepGatewayAnnotation string
	keepDNS               bool
}

func (cfg gatewayPodMutatorCfg) GatewayPodMutator(_ context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		// If not a pod just continue the mutation chain(if there is one) and don't do nothing.
		return &kwhmutating.MutatorResult{}, nil
	}

	//Check if label excludes this pod
	if val, ok := pod.GetLabels()[cfg.keepGatewayLabel]; cfg.keepGatewayLabel != "" && ok {
		if val, err := strconv.ParseBool(val); err == nil && val {
			return &kwhmutating.MutatorResult{
				MutatedObject: pod,
			}, nil
		}
	}

	//Check if annotations excludes this pod
	if val, ok := pod.GetAnnotations()[cfg.keepGatewayAnnotation]; cfg.keepGatewayAnnotation != "" && ok {
		if val, err := strconv.ParseBool(val); err == nil && val {
			return &kwhmutating.MutatorResult{
				MutatedObject: pod,
			}, nil
		}
	}

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

	return &kwhmutating.MutatorResult{
		MutatedObject: pod,
	}, nil

}
