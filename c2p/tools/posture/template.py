# -*- mode:python; coding:utf-8 -*-

# Copyright 2026 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
"""Markdown view over a c2p-posture/v1 document.

Every control the component definition expects gets a row, including the ones nothing
assessed. A view that lists only controls with results cannot show the interesting case.
"""

TEMPLATE = """# Compliance Posture
{% if posture.framework %}
Framework: {{ posture.framework }}
{% endif %}
{{ posture.summary.satisfied }} satisfied, {{ posture.summary.not_satisfied }} not satisfied, \
{{ posture.summary.not_assessed }} not assessed
{% for component in posture.components %}
## Component: {{ component.title }}

| Control | Status | Coverage | Checks | Notes |
|---|---|---|---|---|
{%- for control in component.controls %}
| {{ control.control_id }}{% if control.title %}<br/>{{ control.title }}{% endif %} \
| {{ icon(control.state) }} {{ control.state }} \
| {{ control.coverage.assessed_checks }}/{{ control.coverage.expected_checks }} \
| {% for check in flatten_checks(control.rules) %}{{ icon(check.result) }} {{ check.engine }}\
{% if not loop.last %}<br/>{% endif %}{% endfor %} \
| {{ control.remarks }} |
{%- endfor %}
{% endfor %}"""

# The HTML view carries what a table cannot: every subject, with the engine's own message.
# A posture that shows only a colour per control is not auditable -- the reason is the part a
# reader has to be able to disagree with.
HTML_TEMPLATE = """<!doctype html>
<meta charset="utf-8">
<title>Compliance posture</title>
<style>
body{font-family:system-ui,sans-serif;margin:2rem;color:#1c1c1c;background:#fafafa}
/* Fixed layout, or the subjects column takes the width and squeezes the rule column until
   every hyphenated check id wraps one word per line. */
table{border-collapse:collapse;width:100%;background:#fff;table-layout:fixed}
col.control{width:14%}col.state{width:13%}col.coverage{width:6%}col.checks{width:27%}
th,td{border:1px solid #ddd;padding:.45rem .6rem;vertical-align:top;font-size:.92rem}
th{background:#f0f0f0;text-align:left}
td{overflow-wrap:anywhere}
.satisfied{background:#e6f7e9}.not-satisfied{background:#fde8e8}.not-assessed{background:#fff6d6}
code{font-size:.85em;white-space:nowrap}ul{margin:.1rem 0;padding-left:1.1rem}li{margin:.15rem 0}
/* The subject list is the audit trail and has to be present, but a control with 34 subjects
   makes the table unreadable if they are all open: one control then fills a screen and the
   posture -- which is the point of the page -- cannot be scanned at all. */
details>summary{cursor:pointer;font-size:.9rem}
details[open]>summary{margin-bottom:.3rem}
h1{margin-bottom:0}.meta{color:#666;margin-top:.2rem}
</style>
<h1>Compliance posture</h1>
<p class="meta">
{% if posture.framework %}{{ posture.framework }} · {% endif %}
{{ posture.summary.satisfied }} satisfied · {{ posture.summary.not_satisfied }} not satisfied ·
{{ posture.summary.not_assessed }} not assessed
{% if posture.generated_at %}· generated {{ posture.generated_at }}{% endif %}
</p>
{% for component in posture.components %}
<h2>{{ component.title }}</h2>
<table>
<col class="control"><col class="state"><col class="coverage"><col class="checks"><col>
<tr><th>Control</th><th>State</th><th>Coverage</th><th>Rule → check (engine)</th><th>Subjects</th></tr>
{%- for control in component.controls %}
<tr class="{{ control.state }}">
  <td><b>{{ control.control_id }}</b>{% if control.title %}<br><small>{{ control.title }}</small>{% endif %}</td>
  <td>{{ control.state }}<br><small>{{ control.remarks }}</small></td>
  <td>{{ control.coverage.assessed_checks }}/{{ control.coverage.expected_checks }}</td>
  <td>
  {%- for rule in control.rules %}
    {%- for check in rule.checks %}
    <code>{{ rule.rule_id }}</code> → <code>{{ check.check_id if check.check_id else '—' }}</code>
    ({{ check.engine }}) <b>{{ check.result }}</b><br>
    {%- endfor %}
  {%- endfor %}
  </td>
  <td>
  {%- for rule in control.rules %}{% for check in rule.checks %}
    {%- if check.subjects %}
    <details>
    <summary><code>{{ check.check_id if check.check_id else rule.rule_id }}</code>
    — {{ check.subjects | length }} subject(s),
    {{ check.subjects | rejectattr('result', 'equalto', 'pass') | list | length }} not passing</summary>
    <ul>
    {%- for subject in check.subjects %}
      <li><b>{{ subject.result }}</b> {{ subject.title }}
      {%- if subject.reason %}<br><small>{{ subject.reason }}</small>{% endif %}</li>
    {%- endfor %}
    </ul>
    </details>
    {%- else %}
    <ul><li><i>{{ check.remarks }}</i></li></ul>
    {%- endif %}
  {%- endfor %}{% endfor %}
  </td>
</tr>
{%- endfor %}
</table>
{% endfor %}
<p class="meta">A denied non-compliant request is recorded as a pass: the control was effective.
<b>not-assessed</b> means no engine could evaluate the control -- neither compliant nor
violated, and not the same as passing.</p>
"""
