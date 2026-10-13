package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeDiffWithRedactionMasksSensitiveAndConfiguredFields(t *testing.T) {
	oldObject := map[string]interface{}{
		"data": map[string]interface{}{
			"apiToken":      "old-token-value",
			"privateConfig": "old-private-value",
		},
		"visible": "old",
	}
	newObject := map[string]interface{}{
		"data": map[string]interface{}{
			"apiToken":      "new-token-value",
			"privateConfig": "new-private-value",
		},
		"visible": "new",
	}

	got := ComputeDiffWithRedaction(oldObject, newObject, []string{"data.privateConfig"})

	assert.Equal(t,
		"- visible: old\n+ visible: new\n~ data.apiToken: [REDACTED] (changed)\n~ data.privateConfig: [REDACTED] (changed)\n",
		got,
	)
	assert.NotContains(t, got, "old-token-value")
	assert.NotContains(t, got, "new-token-value")
	assert.NotContains(t, got, "old-private-value")
	assert.NotContains(t, got, "new-private-value")
}

func TestComputeDiffWithRedactionMasksSensitiveEnvironmentVariableValues(t *testing.T) {
	oldObject := map[string]interface{}{
		"containers": []interface{}{
			map[string]interface{}{
				"name": "app",
				"env": []interface{}{
					map[string]interface{}{"name": "API_TOKEN", "value": "old-token-value"},
				},
			},
		},
	}
	newObject := map[string]interface{}{
		"containers": []interface{}{
			map[string]interface{}{
				"name": "app",
				"env": []interface{}{
					map[string]interface{}{"name": "API_TOKEN", "value": "new-token-value"},
				},
			},
		},
	}

	got := ComputeDiffWithRedaction(oldObject, newObject, nil)

	assert.Equal(t, "~ containers.0.env.0.value: [REDACTED] (changed)\n", got)
	assert.NotContains(t, got, "old-token-value")
	assert.NotContains(t, got, "new-token-value")
}

func TestComputeDiffWithRedactionSupportsWildcardPaths(t *testing.T) {
	oldObject := map[string]interface{}{"data": map[string]interface{}{"first": "old-first", "second": "old-second"}}
	newObject := map[string]interface{}{"data": map[string]interface{}{"first": "new-first", "second": "new-second"}}

	got := ComputeDiffWithRedaction(oldObject, newObject, []string{"data.*"})

	assert.Equal(t, "~ data.first: [REDACTED] (changed)\n~ data.second: [REDACTED] (changed)\n", got)
	assert.NotContains(t, got, "old-first")
	assert.NotContains(t, got, "new-first")
}

func TestComputeDiffWithRedactionSupportsEscapedMapKeys(t *testing.T) {
	oldObject := map[string]interface{}{
		"data": map[string]interface{}{"app.yaml": "old-sensitive-value", "other": "old"},
	}
	newObject := map[string]interface{}{
		"data": map[string]interface{}{"app.yaml": "new-sensitive-value", "other": "new"},
	}

	got := ComputeDiffWithRedaction(oldObject, newObject, []string{`data.app\.yaml`})

	assert.Equal(t,
		"- data: map[other:old]\n+ data: map[other:new]\n~ data.app\\.yaml: [REDACTED] (changed)\n",
		got,
	)
	assert.NotContains(t, got, "old-sensitive-value")
	assert.NotContains(t, got, "new-sensitive-value")
}

func TestComputeDiffWithRedactionDoesNotReportUnchangedSensitiveFields(t *testing.T) {
	oldObject := map[string]interface{}{"apiToken": "same-token-value"}
	newObject := map[string]interface{}{"apiToken": "same-token-value"}

	assert.Empty(t, ComputeDiffWithRedaction(oldObject, newObject, nil))
}

func TestComputeDiffWithRedactionDoesNotMutateInputs(t *testing.T) {
	oldObject := map[string]interface{}{"credentials": map[string]interface{}{"password": "old-password"}}
	newObject := map[string]interface{}{"credentials": map[string]interface{}{"password": "new-password"}}
	oldCopy := map[string]interface{}{"credentials": map[string]interface{}{"password": "old-password"}}
	newCopy := map[string]interface{}{"credentials": map[string]interface{}{"password": "new-password"}}

	ComputeDiffWithRedaction(oldObject, newObject, nil)

	assert.Equal(t, oldCopy, oldObject)
	assert.Equal(t, newCopy, newObject)
}
