package diff

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// ComputeDiff calculates the differences between two maps
func ComputeDiff(oldMap, newMap map[string]interface{}) string {
	var diff string
	newKeys := make([]string, 0, len(newMap))
	for key := range newMap {
		newKeys = append(newKeys, key)
	}
	sort.Strings(newKeys)

	// Compare keys in the new map
	for _, key := range newKeys {
		newVal := newMap[key]
		if oldVal, exists := oldMap[key]; exists {
			// Key exists in both maps
			if reflect.TypeOf(oldVal) == reflect.TypeOf(newVal) {
				switch oldValTyped := oldVal.(type) {
				case map[string]interface{}:
					// Recursive call for nested maps
					newValTyped := newVal.(map[string]interface{})
					nestedDiff := ComputeDiff(oldValTyped, newValTyped)
					if nestedDiff != "" {
						diff += fmt.Sprintf("%s:\n%s", key, nestedDiff)
					}
				default:
					// Compare leaf values
					if !reflect.DeepEqual(oldVal, newVal) {
						diff += fmt.Sprintf("- %s: %s\n+ %s: %s\n", key, formatValue(oldVal), key, formatValue(newVal))
					}
				}
			} else {
				// Type mismatch
				diff += fmt.Sprintf("- %s: %s\n+ %s: %s\n", key, formatValue(oldVal), key, formatValue(newVal))
			}
		} else {
			// Key added
			diff += fmt.Sprintf("+ %s: %s\n", key, formatValue(newVal))
		}
	}

	// Find removed keys
	oldKeys := make([]string, 0, len(oldMap))
	for key := range oldMap {
		oldKeys = append(oldKeys, key)
	}
	sort.Strings(oldKeys)
	for _, key := range oldKeys {
		oldVal := oldMap[key]
		if _, exists := newMap[key]; !exists {
			// Key removed
			diff += fmt.Sprintf("- %s: %s\n", key, formatValue(oldVal))
		}
	}

	// Format the diff output for better readability
	var formattedDiff bytes.Buffer
	for _, line := range bytes.Split([]byte(diff), []byte("\n")) {
		if bytes.HasPrefix(line, []byte("+")) || bytes.HasPrefix(line, []byte("-")) {
			formattedDiff.Write(line)
			formattedDiff.WriteByte('\n')
		}
	}

	return formattedDiff.String()
}

// formatValue keeps multiline strings on one diff line so their content is not
// mistaken for unprefixed diff output lines.
func formatValue(value interface{}) string {
	if text, ok := value.(string); ok && strings.ContainsAny(text, "\r\n") {
		return strconv.Quote(text)
	}
	return fmt.Sprint(value)
}

// RemoveFields removes specified fields from a map
func RemoveFields(obj map[string]interface{}, fields []string) map[string]interface{} {
	for _, field := range fields {
		if field == "" {
			continue
		}

		parts := strings.Split(field, ".")
		removeNestedField(obj, parts)
	}
	return obj
}

// removeNestedField removes a nested field from a map
func removeNestedField(obj map[string]interface{}, parts []string) {
	if len(parts) == 1 {
		delete(obj, parts[0])
		return
	}

	if nestedMap, exists := obj[parts[0]].(map[string]interface{}); exists {
		removeNestedField(nestedMap, parts[1:])
	}
}
