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

	"github.com/k8s-at-home/gateway-admision-controller/internal/config"
	mutator "github.com/k8s-at-home/gateway-admision-controller/internal/mutation"
)

const (
	testSidecarImage      = "foo"
	testSidecarCmd        = "bla"
	testSidecarMountPoint = "/mnt"
	testSidecarConfigmap  = "settings"
	testDNSPolicy         = "None"
)

func getExpectedPodSpec(gatewayIP string) corev1.PodSpec {
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
		DNSConfig: &corev1.PodDNSConfig{
			Nameservers: []string{
				gatewayIP,
			},
		},
	}
	return spec
}

func getExpectedPodSpec_keepDNS(gatewayIP string) corev1.PodSpec {
	spec := getExpectedPodSpec(gatewayIP)
	spec.DNSPolicy = ""
	spec.DNSConfig = nil
	return spec
}

func getExpectedPodSpec_sidecar(gatewayIP string) corev1.PodSpec {
	spec := getExpectedPodSpec(gatewayIP)

	container := corev1.Container{
		Name:    "gateway-sidecar",
		Image:   testSidecarImage,
		Command: []string{testSidecarCmd},
		Env: []corev1.EnvVar{
			{
				Name:  "gateway",
				Value: gatewayIP,
			},
		},
		SecurityContext: &corev1.SecurityContext{
			Privileged: &[]bool{true}[0],
		},
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      mutator.GATEWAY_SIDECAR_VOLUME_NAME,
				MountPath: testSidecarMountPoint,
			},
		},
	}

	//Add container to pod
	spec.Containers = append(spec.Containers, container)

	spec.Volumes = append(spec.Volumes, corev1.Volume{
		Name: mutator.GATEWAY_SIDECAR_VOLUME_NAME,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: testSidecarConfigmap,
				},
				DefaultMode: &mutator.GATEWAY_SIDECAR_VOLUME_MODE,
			},
		},
	})

	return spec
}

func getExpectedPodSpec_DNSPolicy(gatewayIP string) corev1.PodSpec {
	spec := getExpectedPodSpec(gatewayIP)
	spec.DNSPolicy = testDNSPolicy
	return spec
}

func TestGatewayPodMutator(t *testing.T) {

	exampleGatewayName := "example.com"
	exampleGatewayNameIPs, _ := net.LookupIP(exampleGatewayName)

	tests := map[string]struct {
		cmdConfig config.CmdConfig
		obj       metav1.Object
		expObj    metav1.Object
	}{
		"Gateway IP - Having a pod, gateway should be added": {
			cmdConfig: config.CmdConfig{
				Gateway:           "1.2.3.4",
				SetGatewayDefault: true,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec("1.2.3.4"),
			},
		},
		"Gateway name - Having a pod, gateway should be added": {
			cmdConfig: config.CmdConfig{
				Gateway:           exampleGatewayName,
				SetGatewayDefault: true,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec(exampleGatewayNameIPs[0].String()),
			},
		},
		"Gateway IP, keepDNS=true - Having a pod, gateway should be added": {
			cmdConfig: config.CmdConfig{
				Gateway:           "1.2.3.4",
				SetGatewayDefault: true,
				KeepDNS:           true,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec_keepDNS("1.2.3.4"),
			},
		},
		"Gateway IP, no SetGatewayDefault - it should be a NOP": {
			cmdConfig: config.CmdConfig{
				Gateway: "1.2.3.4",
			},
			obj:    &corev1.Pod{},
			expObj: &corev1.Pod{},
		},
		"Gateway IP, setGatewayLabel='setGateway' - it should be a NOP": {
			cmdConfig: config.CmdConfig{
				Gateway:           "1.2.3.4",
				SetGatewayDefault: true,
				SetGatewayLabel:   "setGateway",
			},
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"setGateway": "false",
					},
				},
			},
			expObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"setGateway": "false",
					},
				},
			},
		},
		"Gateway IP, setGatewayLabel='setGateway' - it should set gateway since label is true": {
			cmdConfig: config.CmdConfig{
				Gateway:           "1.2.3.4",
				SetGatewayDefault: true,
				SetGatewayLabel:   "setGateway",
			},
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"setGateway": "true",
					},
				},
			},
			expObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"setGateway": "true",
					},
				},
				Spec: getExpectedPodSpec("1.2.3.4"),
			},
		},
		"Gateway IP, setGatewayAnnotation='setGateway' - it should be a NOP": {
			cmdConfig: config.CmdConfig{
				Gateway:              "1.2.3.4",
				SetGatewayDefault:    true,
				SetGatewayAnnotation: "setGateway",
			},
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"setGateway": "false",
					},
				},
			},
			expObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"setGateway": "false",
					},
				},
			},
		},
		"Gateway IP, setGatewayAnnotation='setGateway' - it should set gateway since label is true": {
			cmdConfig: config.CmdConfig{
				Gateway:              "1.2.3.4",
				SetGatewayDefault:    true,
				SetGatewayAnnotation: "setGateway",
			},
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"setGateway": "true",
					},
				},
			},
			expObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"setGateway": "true",
					},
				},
				Spec: getExpectedPodSpec("1.2.3.4"),
			},
		},
		"Gateway IP, sidecar cmd": {
			cmdConfig: config.CmdConfig{
				Gateway:           "1.2.3.4",
				SetGatewayDefault: true,
				SidecarImage:      testSidecarImage,
				SidecarCmd:        testSidecarCmd,
				SidecarMountPoint: testSidecarMountPoint,
				SidecarConfigmap:  testSidecarConfigmap,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec_sidecar("1.2.3.4"),
			},
		},
		"Gateway IP, DNSPolicy": {
			cmdConfig: config.CmdConfig{
				Gateway:           "1.2.3.4",
				SetGatewayDefault: true,
				SetDNSPolicy:      testDNSPolicy,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec_DNSPolicy("1.2.3.4"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			m, err := mutator.NewGatewayPodMutator(test.cmdConfig)
			require.NoError(err)

			_, err = m.GatewayPodMutator(context.TODO(), nil, test.obj)
			require.NoError(err)

			assert.Equal(test.expObj, test.obj)
		})
	}
}
