package helpers

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// TestPod creates a test pod for testing purposes
func CreateTestPod(name, namespace, image string) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			ResourceVersion:   "1000",
			UID:               "test-uid-12345",
			CreationTimestamp: metav1.Now(),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: image,
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
}

// CreateTestDeployment creates a test deployment for testing purposes
func CreateTestDeployment(name, namespace, image string, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			ResourceVersion:   "2000",
			UID:               "test-deployment-uid-67890",
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: image,
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Replicas:      replicas,
			ReadyReplicas: replicas,
		},
	}
}

// CreateTestService creates a test service for testing purposes
func CreateTestService(name, namespace string, port int32) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			ResourceVersion:   "3000",
			UID:               "test-service-uid-11111",
			CreationTimestamp: metav1.Now(),
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []corev1.ServicePort{
				{
					Port:     port,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// ConvertToUnstructured converts a Kubernetes object to unstructured format
func ConvertToUnstructured(t *testing.T, obj runtime.Object) *unstructured.Unstructured {
	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	require.NoError(t, err)
	return &unstructured.Unstructured{Object: unstructuredMap}
}

// CloneUnstructured creates a deep copy of an unstructured object
func CloneUnstructured(t *testing.T, obj *unstructured.Unstructured) *unstructured.Unstructured {
	data, err := json.Marshal(obj.Object)
	require.NoError(t, err)

	var clonedMap map[string]interface{}
	err = json.Unmarshal(data, &clonedMap)
	require.NoError(t, err)

	return &unstructured.Unstructured{Object: clonedMap}
}

// UpdatePodImage updates the image of a pod
func UpdatePodImage(pod *corev1.Pod, newImage string) *corev1.Pod {
	updatedPod := pod.DeepCopy()
	updatedPod.Spec.Containers[0].Image = newImage
	updatedPod.ResourceVersion = "1001"
	return updatedPod
}

// UpdateDeploymentReplicas updates the replica count of a deployment
func UpdateDeploymentReplicas(deployment *appsv1.Deployment, replicas int32) *appsv1.Deployment {
	updatedDeployment := deployment.DeepCopy()
	updatedDeployment.Spec.Replicas = &replicas
	updatedDeployment.ResourceVersion = "2001"
	return updatedDeployment
}

// CreateTestNamespace creates a test namespace
func CreateTestNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			ResourceVersion:   "4000",
			UID:               types.UID(fmt.Sprintf("test-namespace-uid-%s", name)),
			CreationTimestamp: metav1.Now(),
		},
		Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceActive,
		},
	}
}

// AssertNoError is a helper to assert no error occurred
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	require.NoError(t, err, msgAndArgs...)
}

// AssertError is a helper to assert an error occurred
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	require.Error(t, err, msgAndArgs...)
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, str, substr string, msgAndArgs ...interface{}) {
	require.Contains(t, str, substr, msgAndArgs...)
}

// Int32Ptr returns a pointer to an int32
func Int32Ptr(i int32) *int32 {
	return &i
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}
