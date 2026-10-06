{{- define "k8s-diff-informer.validate" -}}
{{- if and (.Values.slack.webhookUrl | trim) (.Values.slack.existingSecret | trim) -}}
{{- fail "Set only one of slack.webhookUrl or slack.existingSecret" -}}
{{- end -}}
{{- if not (or (.Values.slack.webhookUrl | trim) (.Values.slack.existingSecret | trim)) -}}
{{- fail "Set slack.webhookUrl or slack.existingSecret before installing" -}}
{{- end -}}
{{- end -}}
