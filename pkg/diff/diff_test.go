package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeDiff(t *testing.T) {
	testCases := []struct {
		name     string
		oldMap   map[string]interface{}
		newMap   map[string]interface{}
		expected string
	}{
		{
			name:     "no changes",
			oldMap:   map[string]interface{}{"key": "value"},
			newMap:   map[string]interface{}{"key": "value"},
			expected: "",
		},
		{
			name:     "value changed",
			oldMap:   map[string]interface{}{"key": "old_value"},
			newMap:   map[string]interface{}{"key": "new_value"},
			expected: "- key: old_value\n+ key: new_value\n",
		},
		{
			name:     "key added",
			oldMap:   map[string]interface{}{},
			newMap:   map[string]interface{}{"new_key": "value"},
			expected: "+ new_key: value\n",
		},
		{
			name:     "key removed",
			oldMap:   map[string]interface{}{"removed_key": "value"},
			newMap:   map[string]interface{}{},
			expected: "- removed_key: value\n",
		},
		//{
		//	name: "nested map changes",
		//	oldMap: map[string]interface{}{
		//		"parent": map[string]interface{}{
		//			"child": "old_value",
		//		},
		//	},
		//	newMap: map[string]interface{}{
		//		"parent": map[string]interface{}{
		//			"child": "new_value",
		//		},
		//	},
		//	expected: "parent:\n- child: old_value\n+ child: new_value\n",
		//},
		{
			name: "type mismatch",
			oldMap: map[string]interface{}{
				"key": "string_value",
			},
			newMap: map[string]interface{}{
				"key": 123,
			},
			expected: "- key: string_value\n+ key: 123\n",
		},
		//{
		//	name: "complex nested structure",
		//	oldMap: map[string]interface{}{
		//		"spec": map[string]interface{}{
		//			"replicas": 2,
		//			"template": map[string]interface{}{
		//				"spec": map[string]interface{}{
		//					"containers": []interface{}{
		//						map[string]interface{}{
		//							"name":  "app",
		//							"image": "nginx:1.14",
		//						},
		//					},
		//				},
		//			},
		//		},
		//	},
		//	newMap: map[string]interface{}{
		//		"spec": map[string]interface{}{
		//			"replicas": 3,
		//			"template": map[string]interface{}{
		//				"spec": map[string]interface{}{
		//					"containers": []interface{}{
		//						map[string]interface{}{
		//							"name":  "app",
		//							"image": "nginx:1.15",
		//						},
		//					},
		//				},
		//			},
		//		},
		//	},
		//	expected: "spec:\n- replicas: 2\n+ replicas: 3\ntemplate:\nspec:\n- containers: [map[image:nginx:1.14 name:app]]\n+ containers: [map[image:nginx:1.15 name:app]]\n",
		//},
		{
			name: "nested map changes",
			oldMap: map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"image":           "nginx",
						"imagePullPolicy": "Always",
						"name":            "nginx3",
						"ports": []interface{}{
							map[string]interface{}{
								"containerPort": 80,
								"protocol":      "TCP",
							},
						},
						"resources":                map[string]interface{}{},
						"terminationMessagePath":   "/dev/termination-log",
						"terminationMessagePolicy": "File",
					},
				},
			},
			newMap: map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"image":           "nginx",
						"imagePullPolicy": "Always",
						"name":            "nginxlast",
						"ports": []interface{}{
							map[string]interface{}{
								"containerPort": 80,
								"protocol":      "TCP",
							},
						},
						"resources":                map[string]interface{}{},
						"terminationMessagePath":   "/dev/termination-log",
						"terminationMessagePolicy": "File",
					},
				},
			},
			expected: "- containers: [map[image:nginx imagePullPolicy:Always name:nginx3 ports:[map[containerPort:80 protocol:TCP]] resources:map[] terminationMessagePath:/dev/termination-log terminationMessagePolicy:File]]\n+ containers: [map[image:nginx imagePullPolicy:Always name:nginxlast ports:[map[containerPort:80 protocol:TCP]] resources:map[] terminationMessagePath:/dev/termination-log terminationMessagePolicy:File]]\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ComputeDiff(tc.oldMap, tc.newMap)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRemoveFields(t *testing.T) {
	testCases := []struct {
		name           string
		inputMap       map[string]interface{}
		fieldsToRemove []string
		expectedMap    map[string]interface{}
	}{
		{
			name: "remove top-level field",
			inputMap: map[string]interface{}{
				"keep":   "value1",
				"remove": "value2",
			},
			fieldsToRemove: []string{"remove"},
			expectedMap: map[string]interface{}{
				"keep": "value1",
			},
		},
		{
			name: "remove nested field",
			inputMap: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      "test-pod",
					"namespace": "default",
					"uid":       "12345",
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{},
				},
			},
			fieldsToRemove: []string{"metadata.uid"},
			expectedMap: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      "test-pod",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{},
				},
			},
		},
		{
			name: "remove multiple fields",
			inputMap: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":              "test-pod",
					"namespace":         "default",
					"uid":               "12345",
					"creationTimestamp": "2023-01-01T00:00:00Z",
					"resourceVersion":   "1000",
				},
				"status": map[string]interface{}{
					"phase": "Running",
				},
			},
			fieldsToRemove: []string{
				"metadata.uid",
				"metadata.creationTimestamp",
				"metadata.resourceVersion",
				"status",
			},
			expectedMap: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      "test-pod",
					"namespace": "default",
				},
			},
		},
		{
			name: "remove non-existent field",
			inputMap: map[string]interface{}{
				"keep": "value",
			},
			fieldsToRemove: []string{"nonexistent"},
			expectedMap: map[string]interface{}{
				"keep": "value",
			},
		},
		{
			name: "remove empty field name",
			inputMap: map[string]interface{}{
				"keep": "value",
			},
			fieldsToRemove: []string{""},
			expectedMap: map[string]interface{}{
				"keep": "value",
			},
		},
		{
			name: "deeply nested field removal",
			inputMap: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": map[string]interface{}{
							"keep":   "value1",
							"remove": "value2",
						},
					},
				},
			},
			fieldsToRemove: []string{"level1.level2.level3.remove"},
			expectedMap: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": map[string]interface{}{
							"keep": "value1",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Make a copy of the input map to avoid modifying the original
			//create a shallow copy of the map before passing it to the function
			inputCopy := make(map[string]interface{})
			for k, v := range tc.inputMap {
				inputCopy[k] = v
			}

			result := RemoveFields(inputCopy, tc.fieldsToRemove)
			assert.Equal(t, tc.expectedMap, result)
		})
	}
}

func TestRemoveNestedField(t *testing.T) {
	testCases := []struct {
		name     string
		inputMap map[string]interface{}
		parts    []string
		expected map[string]interface{}
	}{
		{
			name: "single level removal",
			inputMap: map[string]interface{}{
				"keep":   "value1",
				"remove": "value2",
			},
			parts: []string{"remove"},
			expected: map[string]interface{}{
				"keep": "value1",
			},
		},
		{
			name: "nested field removal",
			inputMap: map[string]interface{}{
				"parent": map[string]interface{}{
					"keep":   "value1",
					"remove": "value2",
				},
			},
			parts: []string{"parent", "remove"},
			expected: map[string]interface{}{
				"parent": map[string]interface{}{
					"keep": "value1",
				},
			},
		},
		{
			name: "non-existent parent",
			inputMap: map[string]interface{}{
				"existing": "value",
			},
			parts: []string{"nonexistent", "child"},
			expected: map[string]interface{}{
				"existing": "value",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			removeNestedField(tc.inputMap, tc.parts)
			assert.Equal(t, tc.expected, tc.inputMap)
		})
	}
}

// Benchmark tests
func BenchmarkComputeDiff(b *testing.B) {
	oldMap := map[string]interface{}{
		"spec": map[string]interface{}{
			"replicas": 3,
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "app",
							"image": "nginx:1.14",
						},
					},
				},
			},
		},
	}

	newMap := map[string]interface{}{
		"spec": map[string]interface{}{
			"replicas": 5,
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "app",
							"image": "nginx:1.15",
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeDiff(oldMap, newMap)
	}
}

func BenchmarkRemoveFields(b *testing.B) {
	testMap := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":              "test-pod",
			"namespace":         "default",
			"creationTimestamp": "2023-01-01T00:00:00Z",
			"uid":               "12345",
			"resourceVersion":   "1000",
		},
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{
					"name":  "nginx",
					"image": "nginx:1.14",
				},
			},
		},
		"status": map[string]interface{}{
			"phase": "Running",
		},
	}

	fieldsToRemove := []string{
		"metadata.creationTimestamp",
		"metadata.uid",
		"metadata.resourceVersion",
		"status",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a copy for each iteration
		testMapCopy := make(map[string]interface{})
		for k, v := range testMap {
			testMapCopy[k] = v
		}
		RemoveFields(testMapCopy, fieldsToRemove)
	}
}

func TestComputeDiffEscapesMultilineStringValues(t *testing.T) {
	oldMap := map[string]interface{}{"config.yaml": "server:\n  port: 8080\n"}
	newMap := map[string]interface{}{"config.yaml": "server:\n  port: 9090\n"}

	result := ComputeDiff(oldMap, newMap)

	assert.Contains(t, result, `- config.yaml: "server:\n  port: 8080\n"`)
	assert.Contains(t, result, `+ config.yaml: "server:\n  port: 9090\n"`)
}
