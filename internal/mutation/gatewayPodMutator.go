package gatewayPodMutator

import (
	"context"
	"net"
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
func NewGatewayPodMutator(gateway string, keepDNS bool) (GatewayPodMutator, error) {
	gatewayIPs, error := net.LookupIP(gateway)
	if error != nil {
		return nil, error
	}
	return gatewayPodMutatorCfg{gatewayIPs: gatewayIPs, keepDNS: keepDNS}, nil
}

type gatewayPodMutatorCfg struct {
	gatewayIPs []net.IP
	keepDNS    bool
}

func (cfg gatewayPodMutatorCfg) GatewayPodMutator(_ context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		// If not a pod just continue the mutation chain(if there is one) and don't do nothing.
		return &kwhmutating.MutatorResult{}, nil
	}

	// Create init container
	container := corev1.Container{
		Name:  "addGateway",
		Image: "alpine",
		Command: append(
			strings.Split("ip route add default via", " "),
			cfg.gatewayIPs[0].String(),
		),
		// Args:                     []string{},
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
