package gatewayPodMutator_test

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k8s-at-home/gateway-admision-controller/internal/config"
	"github.com/k8s-at-home/gateway-admision-controller/internal/log"
	mutator "github.com/k8s-at-home/gateway-admision-controller/internal/mutation"
	"github.com/k8s-at-home/gateway-admision-controller/internal/resolv"
)

const (
	testGatewayIP           = "1.2.3.4"
	testGatewayName         = "example.com"
	testDNSIP               = "5.6.7.8"
	testDNSName             = "www.example.com"
	testDNSPolicy           = "None"
	testInitImage           = "initImg"
	testInitImagePullPol    = "Always"
	testInitCmd             = "initCmd"
	testInitMountPoint      = "/media"
	testSidecarImage        = "sidecarImg"
	testSidecarImagePullPol = "IfNotPresent"
	testSidecarCmd          = "sidecarCmd"
	testSidecarMountPoint   = "/mnt"
	testConfigmapName       = "settings"
	testNamespace           = "myNameSpace"
)

func getExpectedPodSpec_gateway(gateway string, DNS string, initImage string, sidecarImage string) corev1.PodSpec {

	var DNS_IP string
	if DNS != "" {
		DNS_IP_obj, _ := net.LookupIP(DNS)
		DNS_IP = DNS_IP_obj[0].String()
	}

	k8s_DNS_config, _ := resolv.Config()
	k8s_DNS_ips := strings.Join(k8s_DNS_config.Nameservers, " ")

	//fix the first search to match the pod namespace
	for i := range k8s_DNS_config.Search {
		searchParts := strings.Split(k8s_DNS_config.Search[i], ".")
		if len(searchParts) > 2 && searchParts[1] == "svc" {
			searchParts[0] = testNamespace
			k8s_DNS_config.Search[i] = strings.Join(searchParts, ".")
		}
	}

	var initContainers []corev1.Container
	var initContainerRunAsUser = int64(0) // Run init container as root
	var initContainerRunAsNonRoot = false
	if initImage != "" {
		initContainers = append(initContainers, corev1.Container{
			Name:    mutator.GATEWAY_INIT_CONTAINER_NAME,
			Image:   initImage,
			Command: []string{testInitCmd},
			Env: []corev1.EnvVar{
				{
					Name:  "gateway",
					Value: gateway,
				},
				{
					Name:  "DNS",
					Value: DNS,
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
			ImagePullPolicy: corev1.PullPolicy(testInitImagePullPol),
			SecurityContext: &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{
						"NET_ADMIN",
					},
					Drop: []corev1.Capability{},
				},
				RunAsUser:    &initContainerRunAsUser,
				RunAsNonRoot: &initContainerRunAsNonRoot,
			},
			VolumeMounts: []corev1.VolumeMount{
				corev1.VolumeMount{
					Name:      mutator.GATEWAY_CONFIGMAP_VOLUME_NAME,
					ReadOnly:  true,
					MountPath: testInitMountPoint,
				},
			},
		})
	}

	var containers []corev1.Container
	var sidecarContainerRunAsUser = int64(0) // Run init container as root
	var sidecarContainerRunAsNonRoot = false
	if sidecarImage != "" {
		containers = append(containers, corev1.Container{
			Name:    mutator.GATEWAY_SIDECAR_CONTAINER_NAME,
			Image:   sidecarImage,
			Command: []string{testSidecarCmd},
			Env: []corev1.EnvVar{
				{
					Name:  "gateway",
					Value: gateway,
				},
				{
					Name:  "DNS",
					Value: DNS,
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
			ImagePullPolicy: corev1.PullPolicy(testSidecarImagePullPol),
			SecurityContext: &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{
						"NET_ADMIN",
					},
					Drop: []corev1.Capability{},
				},
				RunAsUser:    &sidecarContainerRunAsUser,
				RunAsNonRoot: &sidecarContainerRunAsNonRoot,
			},
			VolumeMounts: []corev1.VolumeMount{
				corev1.VolumeMount{
					Name:      mutator.GATEWAY_CONFIGMAP_VOLUME_NAME,
					ReadOnly:  true,
					MountPath: testSidecarMountPoint,
				},
			},
		})
	}

	spec := corev1.PodSpec{
		InitContainers: initContainers,
		Containers:     containers,
	}

	if DNS != "" {
		spec.DNSConfig = &corev1.PodDNSConfig{
			Nameservers: []string{DNS_IP},
		}

		if testDNSPolicy == "None" {
			// Copy my own webhook settings
			spec.DNSConfig.Searches = k8s_DNS_config.Search

			podDNSConfigOptions := make([]corev1.PodDNSConfigOption, 0)
			for i := range k8s_DNS_config.Options {
				podDNSConfigOptions = append(podDNSConfigOptions, corev1.PodDNSConfigOption{
					Name:  k8s_DNS_config.Options[i].Name,
					Value: k8s_DNS_config.Options[i].Value,
				})
			}
			spec.DNSConfig.Options = podDNSConfigOptions
		}
	}

	if initImage != "" || sidecarImage != "" {
		spec.Volumes = append(spec.Volumes, corev1.Volume{
			Name: mutator.GATEWAY_CONFIGMAP_VOLUME_NAME,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: testConfigmapName,
					},
					DefaultMode: &mutator.GATEWAY_CONFIGMAP_VOLUME_MODE,
				},
			},
		})
	}
	return spec
}

func getExpectedPodSpec_DNS(DNS string) corev1.PodSpec {
	DNSIPs, _ := net.LookupIP(DNS)
	spec := corev1.PodSpec{
		DNSConfig: &corev1.PodDNSConfig{
			Nameservers: []string{
				DNSIPs[0].String(),
			},
		},
	}
	return spec
}

func getExpectedPodSpec_DNSPolicy(DNSPolicy string) corev1.PodSpec {
	spec := corev1.PodSpec{
		DNSPolicy: corev1.DNSPolicy(DNSPolicy),
	}
	return spec
}

func getExpectedPodSpec_gateway_DNSPolicy(gateway string, DNS string, initImage string, sidecarImage string, DNSPolicy string) corev1.PodSpec {
	spec := getExpectedPodSpec_gateway(gateway, DNS, initImage, sidecarImage)
	spec.DNSPolicy = corev1.DNSPolicy(DNSPolicy)

	return spec
}

func TestGatewayPodMutator(t *testing.T) {

	tests := map[string]struct {
		cmdConfig config.CmdConfig
		obj       metav1.Object
		expObj    metav1.Object
	}{
		"Empty - NOP": {
			cmdConfig: config.CmdConfig{
				SetGatewayDefault: true,
			},
			obj:    &corev1.Pod{},
			expObj: &corev1.Pod{},
		},
		"Gateway IP, no SetGatewayDefault - it should be a NOP": {
			cmdConfig: config.CmdConfig{
				Gateway:          testGatewayIP,
				InitImage:        testInitImage,
				InitCmd:          testInitCmd,
				InitImagePullPol: testInitImagePullPol,
				InitMountPoint:   testInitMountPoint,
				ConfigmapName:    testConfigmapName,
			},
			obj:    &corev1.Pod{},
			expObj: &corev1.Pod{},
		},
		"Gateway IP, init image": {
			cmdConfig: config.CmdConfig{
				SetGatewayDefault: true,
				Gateway:           testGatewayIP,
				InitImage:         testInitImage,
				InitCmd:           testInitCmd,
				InitImagePullPol:  testInitImagePullPol,
				InitMountPoint:    testInitMountPoint,
				ConfigmapName:     testConfigmapName,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec_gateway(testGatewayIP, "", testInitImage, ""),
			},
		},
		"Gateway name, init image": {
			cmdConfig: config.CmdConfig{
				SetGatewayDefault: true,
				Gateway:           testGatewayName,
				InitImage:         testInitImage,
				InitCmd:           testInitCmd,
				InitImagePullPol:  testInitImagePullPol,
				InitMountPoint:    testInitMountPoint,
				ConfigmapName:     testConfigmapName,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec_gateway(testGatewayName, "", testInitImage, ""),
			},
		},
		"Gateway IP, sidecar image": {
			cmdConfig: config.CmdConfig{
				SetGatewayDefault:   true,
				Gateway:             testGatewayIP,
				SidecarImage:        testSidecarImage,
				SidecarCmd:          testSidecarCmd,
				SidecarImagePullPol: testSidecarImagePullPol,
				SidecarMountPoint:   testSidecarMountPoint,
				ConfigmapName:       testConfigmapName,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec_gateway(testGatewayIP, "", "", testSidecarImage),
			},
		},
		"Gateway name, sidecar image": {
			cmdConfig: config.CmdConfig{
				SetGatewayDefault:   true,
				Gateway:             testGatewayName,
				SidecarImage:        testSidecarImage,
				SidecarCmd:          testSidecarCmd,
				SidecarImagePullPol: testSidecarImagePullPol,
				SidecarMountPoint:   testSidecarMountPoint,
				ConfigmapName:       testConfigmapName,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec_gateway(testGatewayName, "", "", testSidecarImage),
			},
		},
		"DNS": {
			cmdConfig: config.CmdConfig{
				SetGatewayDefault: true,
				DNS:               testDNSIP,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec_DNS(testDNSIP),
			},
		},
		"setGatewayLabel='setGateway' - it should be a NOP since label is false": {
			cmdConfig: config.CmdConfig{
				Gateway:           testGatewayIP,
				SetGatewayDefault: true,
				InitImage:         testInitImage,
				InitCmd:           testInitCmd,
				InitImagePullPol:  testInitImagePullPol,
				InitMountPoint:    testInitMountPoint,
				ConfigmapName:     testConfigmapName,
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
		"setGatewayLabel='setGateway' - it should set gateway since label is true": {
			cmdConfig: config.CmdConfig{
				Gateway:           testGatewayIP,
				SetGatewayDefault: true,
				InitImage:         testInitImage,
				InitCmd:           testInitCmd,
				InitImagePullPol:  testInitImagePullPol,
				InitMountPoint:    testInitMountPoint,
				ConfigmapName:     testConfigmapName,
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
				Spec: getExpectedPodSpec_gateway(testGatewayIP, "", testInitImage, ""),
			},
		},
		"setGatewayAnnotation='setGateway' - it should be a NOP since annotation is true": {
			cmdConfig: config.CmdConfig{
				Gateway:              testGatewayIP,
				SetGatewayDefault:    true,
				InitImage:            testInitImage,
				InitCmd:              testInitCmd,
				InitImagePullPol:     testInitImagePullPol,
				InitMountPoint:       testInitMountPoint,
				ConfigmapName:        testConfigmapName,
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
		"setGatewayAnnotation='setGateway' - it should set gateway since annotation is false": {
			cmdConfig: config.CmdConfig{
				Gateway:              testGatewayIP,
				SetGatewayDefault:    true,
				InitImage:            testInitImage,
				InitCmd:              testInitCmd,
				InitImagePullPol:     testInitImagePullPol,
				InitMountPoint:       testInitMountPoint,
				ConfigmapName:        testConfigmapName,
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
				Spec: getExpectedPodSpec_gateway(testGatewayIP, "", testInitImage, ""),
			},
		},
		"DNSPolicy": {
			cmdConfig: config.CmdConfig{
				SetGatewayDefault: true,
				DNSPolicy:         testDNSPolicy,
			},
			obj: &corev1.Pod{},
			expObj: &corev1.Pod{
				Spec: getExpectedPodSpec_DNSPolicy(testDNSPolicy),
			},
		},
		"DNSPolicy, Gateway IP, init image": {
			cmdConfig: config.CmdConfig{
				SetGatewayDefault: true,
				Gateway:           testGatewayIP,
				DNS:               testDNSIP,
				InitImage:         testInitImage,
				InitCmd:           testInitCmd,
				InitImagePullPol:  testInitImagePullPol,
				InitMountPoint:    testInitMountPoint,
				ConfigmapName:     testConfigmapName,
				DNSPolicy:         testDNSPolicy,
			},
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: testNamespace,
				},
			},
			expObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: testNamespace,
				},
				Spec: getExpectedPodSpec_gateway_DNSPolicy(testGatewayIP, testDNSIP, testInitImage, "", testDNSPolicy),
			},
		},
	}

	logrusLog := logrus.New()
	logrusLogEntry := logrus.NewEntry(logrusLog).WithField("app", "gatewayPodMutator Test")

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			m, err := mutator.NewGatewayPodMutator(test.cmdConfig, log.NewLogrus(logrusLogEntry).WithKV(log.KV{"test": name}))
			require.NoError(err)

			_, err = m.GatewayPodMutator(context.TODO(), nil, test.obj)
			require.NoError(err)

			assert.Equal(test.expObj, test.obj)
		})
	}
}
