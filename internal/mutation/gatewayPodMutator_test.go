package gatewayPodMutator_test

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mutator "github.com/k8s-at-home/gateway-admision-controller/internal/mutation"
)

func getExpectedPodSpec_all(
	gatewayIP string,
	keepDNS bool,
) corev1.PodSpec {
	spec := corev1.PodSpec{
		InitContainers: []corev1.Container{
			corev1.Container{
				Name:    "add-gateway",
				Image:   "alpine",
				Command: []string{"ip"},
				Args: append(
					strings.Split("route change default via", " "),
					gatewayIP,
				),
				SecurityContext: &corev1.SecurityContext{
					Privileged: &[]bool{true}[0],
				},
			},
		},
	}
	if !keepDNS {

		spec.DNSPolicy = "None"
		spec.DNSConfig = &corev1.PodDNSConfig{
			Nameservers: []string{
				gatewayIP,
			},
		}
	}
	return spec
}

func getExpectedPodSpec(gatewayIP string) corev1.PodSpec {
	return getExpectedPodSpec_all(gatewayIP, false)
}

func getExpectedPodSpec_keepDNS(gatewayIP string) corev1.PodSpec {
	return getExpectedPodSpec_all(gatewayIP, true)
}

func TestGatewayPodMutator(t *testing.T) {

	exampleGatewayName := "example.com"
	exampleGatewayNameIPs, _ := net.LookupIP(exampleGatewayName)

	tests := map[string]struct {
		gateway               string
		keepGatewayLabel      string
		keepGatewayAnnotation string
		keepDNS               bool
		obj                   metav1.Object
		expObj                metav1.Object
	}{
		"Gateway IP - Having a pod, gateway should be added": {
			gateway: "1.2.3.4",
			obj:     &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec("1.2.3.4"),
			},
		},
		"Gateway name - Having a pod, gateway should be added": {
			gateway: exampleGatewayName,
			obj:     &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec(exampleGatewayNameIPs[0].String()),
			},
		},
		"Gateway IP, keepDNS=true - Having a pod, gateway should be added": {
			gateway: "1.2.3.4",
			keepDNS: true,
			obj:     &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec_keepDNS("1.2.3.4"),
			},
		},
		"Gateway IP, keepGatewayLabel='keepGateway' - it should be a NOP": {
			gateway:          "1.2.3.4",
			keepGatewayLabel: "keepGateway",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"keepGateway": "true",
					},
				},
			},
			expObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"keepGateway": "true",
					},
				},
			},
		},
		"Gateway IP, keepGatewayLabel='keepGateway' - it should set gateway since label is false": {
			gateway:          "1.2.3.4",
			keepGatewayLabel: "keepGateway",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"keepGateway": "false",
					},
				},
			},
			expObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"keepGateway": "false",
					},
				},
				Spec: getExpectedPodSpec("1.2.3.4"),
			},
		},
		"Gateway IP, keepGatewayAnnotation='keepGateway' - it should be a NOP": {
			gateway:               "1.2.3.4",
			keepGatewayAnnotation: "keepGateway",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"keepGateway": "true",
					},
				},
			},
			expObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"keepGateway": "true",
					},
				},
			},
		},
		"Gateway IP, keepGatewayAnnotation='keepGateway' - it should set gateway since label is false": {
			gateway:               "1.2.3.4",
			keepGatewayAnnotation: "keepGateway",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"keepGateway": "false",
					},
				},
			},
			expObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"keepGateway": "false",
					},
				},
				Spec: getExpectedPodSpec("1.2.3.4"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			m, err := mutator.NewGatewayPodMutator(
				test.gateway,
				test.keepGatewayLabel,
				test.keepGatewayAnnotation,
				test.keepDNS,
			)
			require.NoError(err)

			_, err = m.GatewayPodMutator(context.TODO(), nil, test.obj)
			require.NoError(err)

			assert.Equal(test.expObj, test.obj)
		})
	}
}
