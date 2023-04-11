{{/*
Expand the name of the chart.
*/}}
{{- define "mccp.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "mccp.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "mccp.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "mccp.labels" -}}
helm.sh/chart: {{ include "mccp.chart" . }}
{{ include "mccp.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "mccp.selectorLabels" -}}
app.kubernetes.io/name: {{ include "mccp.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
App selector labels
*/}}
{{- define "mccp.appSelectorLabels" -}}
app.kubernetes.io/part-of: weave-gitops
weave.works/app: weave-gitops-enterprise
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "mccp.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "mccp.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "foo" -}}{{- "zing" }}{{- end }}

{{/*
Return the appropriate apiVersion for ingress.
*/}}
{{- define "common.capabilities.ingress.apiVersion" -}}
{{- if semverCompare "<1.14-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "extensions/v1beta1" -}}
{{- else if semverCompare "<1.19-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "networking.k8s.io/v1beta1" -}}
{{- else -}}
{{- print "networking.k8s.io/v1" -}}
{{- end -}}
{{- end -}}

{{/*
Generate backend entry that is compatible with all Kubernetes API versions.
Usage:
include "common.ingress.backend" (dict "serviceName" "backendName" "servicePort" "backendPort" "context" $)

Params:
  - serviceName - String. Name of an existing service backend
  - servicePort - String/Int. Port name (or number) of the service. It will be translated to different yaml depending if it is a string or an integer.
  - context - Dict - Required. The context for the template evaluation.
*/}}
{{- define "common.ingress.backend" -}}
{{- $apiVersion := (include "common.capabilities.ingress.apiVersion" .context) -}}
{{- if or (eq $apiVersion "extensions/v1beta1") (eq $apiVersion "networking.k8s.io/v1beta1") -}}
serviceName: {{ .serviceName }}
servicePort: {{ .servicePort }}
{{- else -}}
service:
  name: {{ .serviceName }}
  port:
    {{- if typeIs "string" .servicePort }}
    name: {{ .servicePort }}
    {{- else if or (typeIs "int" .servicePort) (typeIs "float64" .servicePort) }}
    number: {{ .servicePort | int }}
    {{- end }}
{{- end -}}
{{- end -}}

{{/*
Returns true if the ingressClassname field is supported
Usage:
include "common.ingress.supportsIngressClassname" .
*/}}
{{- define "common.ingress.supportsIngressClassname" -}}
{{- if semverCompare "<1.18-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "false" -}}
{{- else -}}
{{- print "true" -}}
{{- end -}}
{{- end -}}

{{/*
Returns true if the kubernetes.io/ingress.class annotation is supported
Usage:
include "common.ingress.supportsIngressClassAnnotation" .
*/}}
{{- define "common.ingress.supportsIngressClassAnnotation" -}}
{{- if semverCompare "<1.18-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "true" -}}
{{- else -}}
{{- print "false" -}}
{{- end -}}
{{- end -}}

{{/*
Renders a value that contains template.
Usage:
{{ include "common.tplvalues.render" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "common.tplvalues.render" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}
