# -*- mode:python; coding:utf-8 -*-

# Copyright 2024 IBM Corporation
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

TEMPLATE = """
{% for component in components %}
## Component: {{ component.title }}

{% for control_result in component.control_results %}
#### Result of control {{ control_result.id }}: {{ control_result.description }}

{% for rule_result in control_result.rule_results %}
{% if rule_result.subjects|length > 0 %}
Rule `{{ rule_result.id}}`:
- {{ rule_result.description}}

<details><summary>Details</summary>
{% for subject in rule_result.subjects %}

  - Subject UUID: {{ subject.uuid }}
    - Title: {{ subject.title }}
    - Result: {{ subject.result}}
    - Reason:
      ```
      {{ subject.reason }}
      ```
{% endfor %}
</details>
{% else %}
Rule ID: {{ rule_result.id }}
  - No subjects found
{% endif %}
{% endfor %}
---
{% endfor %}
{% endfor %}
"""
