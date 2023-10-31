## Catalog
{{.CatalogTitle}}

{{- range $component := .Components}}
## Component: {{$component.ComponentTitle}}

{{- range $controlResult := $component.ControlResults}}
#### Result of control: {{$controlResult.ControlId}}

{{- range $ruleResult := $controlResult.RuleResults}}
{{ if gt (len $ruleResult.Subjects) 0 }}
Rule ID: {{$ruleResult.RuleId}}
<details><summary>Details</summary>
{{- range $subject := $ruleResult.Subjects}}

  - Subject UUID: {{$subject.UUID}}
    - Title: {{$subject.Title}}
    - Result: {{$subject.Result}}
    - Reason:
      ```
      {{$subject.Reason}}
      ```
{{- end}}
</details>
{{ else }}
Rule ID: {{$ruleResult.RuleId}}
  - No subjects found
{{ end }}
{{- end}}
---
{{- end}}
{{- end}}
