{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "leanix-k8s-connector.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "leanix-k8s-connector.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "leanix-k8s-connector.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "leanix-k8s-connector.labels" -}}
app.kubernetes.io/name: {{ include "leanix-k8s-connector.name" . }}
helm.sh/chart: {{ include "leanix-k8s-connector.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Ensure standard schedule lowest value is every hour
*/}}

{{- define "leanix-k8s-connector.integrationApiSchedule" -}}
{{- if regexMatch "^([0-9]|[1-5][0-9]) (\\*|\\*/\\d+|\\d+) (\\*|\\*/\\d+|\\d+) (\\*|\\*/\\d+|\\d+) (\\*|\\*/\\d+|\\d+)$" .Values.schedule.standard -}}
{{- printf "%s" .Values.schedule.standard -}}
{{- else -}}
{{- printf "%s" "0 */1 * * *" -}}
{{- end -}}
{{- end -}}
