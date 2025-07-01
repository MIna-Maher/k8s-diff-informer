package fixtures

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// PodFixtures contains test pod configurations
var PodFixtures = struct {
	BasicPod       *corev1.Pod
	MultiContainer *corev1.Pod
	WithVolumes    *corev1.Pod
}{
	BasicPod: &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "basic-pod",
			Namespace:       "default",
			ResourceVersion: "1000",
			Labels: map[string]string{
				"app": "test",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx:1.14",
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 80,
						},
					},
				},
			},
		},
	},

	MultiContainer: &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "multi-container-pod",
			Namespace:       "default",
			ResourceVersion: "1001",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx:1.14",
				},
				{
					Name:  "sidecar",
					Image: "busybox:1.28",
				},
			},
		},
	},

	WithVolumes: &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "pod-with-volumes",
			Namespace:       "default",
			ResourceVersion: "1002",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "app",
					Image: "nginx:1.14",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "config",
							MountPath: "/etc/config",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "app-config",
							},
						},
					},
				},
			},
		},
	},
}

// DeploymentFixtures contains test deployment configurations
var DeploymentFixtures = struct {
	BasicDeployment   *appsv1.Deployment
	ScaledDeployment  *appsv1.Deployment
	UpdatedDeployment *appsv1.Deployment
}{
	BasicDeployment: &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "basic-deployment",
			Namespace:       "default",
			ResourceVersion: "2000",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "basic-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "basic-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "nginx:1.14",
						},
					},
				},
			},
		},
	},

	ScaledDeployment: &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "basic-deployment",
			Namespace:       "default",
			ResourceVersion: "2001",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(5),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "basic-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "basic-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "nginx:1.14",
						},
					},
				},
			},
		},
	},

	UpdatedDeployment: &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "basic-deployment",
			Namespace:       "default",
			ResourceVersion: "2002",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "basic-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "basic-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "nginx:1.15", // Updated image
						},
					},
				},
			},
		},
	},
}

// ServiceFixtures contains test service configurations
var ServiceFixtures = struct {
	ClusterIPService *corev1.Service
	NodePortService  *corev1.Service
	LoadBalancer     *corev1.Service
}{
	ClusterIPService: &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "clusterip-service",
			Namespace:       "default",
			ResourceVersion: "3000",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app": "basic-app",
			},
			Ports: []corev1.ServicePort{
				{
					Port:     80,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	},

	NodePortService: &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "nodeport-service",
			Namespace:       "default",
			ResourceVersion: "3001",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": "basic-app",
			},
			Ports: []corev1.ServicePort{
				{
					Port:     80,
					NodePort: 30080,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	},

	LoadBalancer: &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "loadbalancer-service",
			Namespace:       "default",
			ResourceVersion: "3002",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Selector: map[string]string{
				"app": "basic-app",
			},
			Ports: []corev1.ServicePort{
				{
					Port:     80,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	},
}

// ConfigMapFixtures contains test configmap configurations
var ConfigMapFixtures = struct {
	BasicConfigMap *corev1.ConfigMap
	EnvConfigMap   *corev1.ConfigMap
}{
	BasicConfigMap: &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "basic-configmap",
			Namespace:       "default",
			ResourceVersion: "4000",
		},
		Data: map[string]string{
			"config.yaml": `
server:
  port: 8080
  host: localhost
database:
  url: postgres://localhost:5432/mydb
`,
			"app.properties": "debug=true\nlog_level=info",
		},
	},

	EnvConfigMap: &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "env-configmap",
			Namespace:       "default",
			ResourceVersion: "4001",
		},
		Data: map[string]string{
			"DATABASE_URL": "postgres://localhost:5432/mydb",
			"API_KEY":      "test-api-key-12345",
			"DEBUG":        "true",
		},
	},
}

// Helper function
func int32Ptr(i int32) *int32 {
	return &i
}

// ToUnstructured converts any Kubernetes object to unstructured format
func ToUnstructured(obj runtime.Object) (*unstructured.Unstructured, error) {
	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: unstructuredMap}, nil
}
