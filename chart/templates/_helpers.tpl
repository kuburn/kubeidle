{{/*
Generate the fullname for the resources.
*/}}
{{- define "kubeidle.fullname" -}}
{{- printf "%s-%s" .Release.Name (include "kubeidle.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Generate the name for the resources.
*/}}
{{- define "kubeidle.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Helper to generate labels for resources.
*/}}
{{- define "kubeidle.labels" -}}
app.kubernetes.io/name: {{ include "kubeidle.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | default .Chart.Version | quote}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "kubeidle.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kubeidle.name" . }}
{{- end -}}
