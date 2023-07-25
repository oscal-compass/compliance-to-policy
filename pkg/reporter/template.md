## Catalog
{{.Catalog}}

{{- range $component := .Components}}
## Component: {{$component.ComponentTitle}}

Compliance status: {{$component.ComplianceStatus}}

Checked controls: [{{- range $checkedControl := $component.CheckedControls}}{{$checkedControl}},{{- end}}]

{{- range $controlResult := $component.ControlResults}}
#### Result of control: {{$controlResult.ControlId}}
**Compliance status: {{$controlResult.ComplianceStatus}}**

Rules:
{{- range $ruleResult := $controlResult.RuleResults}}
- Rule ID: {{$ruleResult.RuleId}}
- Policy ID: {{$ruleResult.PolicyId}}
- Status: {{$ruleResult.Status}}
- Reason:
```
{{$ruleResult.Reason}}
```
{{- end}}
---
{{- end}}
{{- end}}
