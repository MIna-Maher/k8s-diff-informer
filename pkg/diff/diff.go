package diff

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const redactedValue = "[REDACTED]"

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

// ComputeDiffWithRedaction computes a diff without exposing values whose paths
// are configured for redaction or whose field names look like credentials.
// Redacted changes are reported by path, with their values replaced by a
// marker. The input maps are not modified.
func ComputeDiffWithRedaction(oldMap, newMap map[string]interface{}, redactedFields []string) string {
	safeOld := redactMap(oldMap, "", redactedFields)
	safeNew := redactMap(newMap, "", redactedFields)
	diffText := ComputeDiff(safeOld, safeNew)

	changedFields := make(map[string]struct{})
	collectRedactedChanges(oldMap, newMap, true, true, "", redactedFields, false, changedFields)
	if len(changedFields) == 0 {
		return diffText
	}

	paths := make([]string, 0, len(changedFields))
	for path := range changedFields {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var output strings.Builder
	output.WriteString(diffText)
	for _, path := range paths {
		fmt.Fprintf(&output, "~ %s: %s (changed)\n", path, redactedValue)
	}
	return output.String()
}

func redactMap(input map[string]interface{}, path string, redactedFields []string) map[string]interface{} {
	output := make(map[string]interface{}, len(input))
	environmentValueIsSensitive := false
	if name, ok := input["name"].(string); ok {
		environmentValueIsSensitive = isSensitiveIdentifier(name)
	}

	for key, value := range input {
		fieldPath := joinPath(path, key)
		if isRedactedField(fieldPath, key, redactedFields, environmentValueIsSensitive && key == "value") {
			continue
		}
		output[key] = redactValue(value, fieldPath, redactedFields)
	}
	return output
}

func redactValue(value interface{}, path string, redactedFields []string) interface{} {
	if matchesRedactedPath(path, redactedFields) {
		return redactedValue
	}

	switch typed := value.(type) {
	case map[string]interface{}:
		return redactMap(typed, path, redactedFields)
	case []interface{}:
		output := make([]interface{}, len(typed))
		for index, item := range typed {
			itemPath := joinPath(path, strconv.Itoa(index))
			if matchesRedactedPath(itemPath, redactedFields) {
				output[index] = redactedValue
			} else {
				output[index] = redactValue(item, itemPath, redactedFields)
			}
		}
		return output
	default:
		return value
	}
}

func collectRedactedChanges(oldValue, newValue interface{}, oldExists, newExists bool, path string, redactedFields []string, environmentValueIsSensitive bool, changes map[string]struct{}) {
	if path != "" {
		pathParts := splitPath(path)
		fieldName := pathParts[len(pathParts)-1]
		if isRedactedField(path, fieldName, redactedFields, environmentValueIsSensitive) {
			if oldExists != newExists || !reflect.DeepEqual(oldValue, newValue) {
				changes[path] = struct{}{}
			}
			return
		}
	}

	oldMap, oldIsMap := oldValue.(map[string]interface{})
	newMap, newIsMap := newValue.(map[string]interface{})
	if oldIsMap || newIsMap {
		keys := make(map[string]struct{}, len(oldMap)+len(newMap))
		for key := range oldMap {
			keys[key] = struct{}{}
		}
		for key := range newMap {
			keys[key] = struct{}{}
		}
		orderedKeys := make([]string, 0, len(keys))
		for key := range keys {
			orderedKeys = append(orderedKeys, key)
		}
		sort.Strings(orderedKeys)

		oldEnvSensitive := sensitiveEnvironmentName(oldMap)
		newEnvSensitive := sensitiveEnvironmentName(newMap)
		for _, key := range orderedKeys {
			oldChild, oldChildExists := oldMap[key]
			newChild, newChildExists := newMap[key]
			childPath := joinPath(path, key)
			collectRedactedChanges(
				oldChild,
				newChild,
				oldChildExists,
				newChildExists,
				childPath,
				redactedFields,
				key == "value" && (oldEnvSensitive || newEnvSensitive),
				changes,
			)
		}
		return
	}

	oldList, oldIsList := oldValue.([]interface{})
	newList, newIsList := newValue.([]interface{})
	if oldIsList || newIsList {
		length := len(oldList)
		if len(newList) > length {
			length = len(newList)
		}
		for index := 0; index < length; index++ {
			oldChildExists := index < len(oldList)
			newChildExists := index < len(newList)
			var oldChild, newChild interface{}
			if oldChildExists {
				oldChild = oldList[index]
			}
			if newChildExists {
				newChild = newList[index]
			}
			collectRedactedChanges(
				oldChild,
				newChild,
				oldChildExists,
				newChildExists,
				joinPath(path, strconv.Itoa(index)),
				redactedFields,
				false,
				changes,
			)
		}
	}
}

func sensitiveEnvironmentName(object map[string]interface{}) bool {
	name, ok := object["name"].(string)
	return ok && isSensitiveIdentifier(name)
}

func isRedactedField(path, fieldName string, redactedFields []string, environmentValueIsSensitive bool) bool {
	return matchesRedactedPath(path, redactedFields) || isSensitiveFieldName(fieldName) || environmentValueIsSensitive
}

func isSensitiveFieldName(fieldName string) bool {
	normalized := normalizeIdentifier(fieldName)
	// Secret reference names identify an object but do not contain its value.
	if strings.HasSuffix(normalized, "name") || strings.HasSuffix(normalized, "ref") {
		return false
	}
	return isSensitiveIdentifier(normalized)
}

func isSensitiveIdentifier(value string) bool {
	normalized := normalizeIdentifier(value)
	for _, marker := range []string{
		"password", "passwd", "token", "secret", "credential", "authorization",
		"privatekey", "apikey", "accesskey",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func normalizeIdentifier(value string) string {
	return strings.Map(func(character rune) rune {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			return unicode.ToLower(character)
		}
		return -1
	}, value)
}

func matchesRedactedPath(path string, patterns []string) bool {
	pathParts := splitPath(path)
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		patternParts := splitPath(pattern)
		if len(patternParts) != len(pathParts) {
			continue
		}
		matches := true
		for index, part := range patternParts {
			if part != "*" && part != pathParts[index] {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return false
}

func joinPath(parent, child string) string {
	child = strings.ReplaceAll(child, `\`, `\\`)
	child = strings.ReplaceAll(child, ".", `\.`)
	if parent == "" {
		return child
	}
	return parent + "." + child
}

func splitPath(path string) []string {
	parts := make([]string, 0, strings.Count(path, ".")+1)
	var part strings.Builder
	escaped := false
	for _, character := range path {
		if escaped {
			part.WriteRune(character)
			escaped = false
			continue
		}
		if character == '\\' {
			escaped = true
			continue
		}
		if character == '.' {
			parts = append(parts, part.String())
			part.Reset()
			continue
		}
		part.WriteRune(character)
	}
	if escaped {
		part.WriteRune('\\')
	}
	return append(parts, part.String())
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
