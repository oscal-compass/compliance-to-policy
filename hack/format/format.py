# Copyright 2023 IBM Corporation

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import yaml
import sys
from prettytable import PrettyTable

with open(sys.argv[1], 'r') as file:
  data = yaml.safe_load(file)

items = data['items']
config_policies_per_cluster = {}
for item in items:
  cluster = item['metadata']['annotations']['kcp.io/cluster']
  if cluster in config_policies_per_cluster:
    config_policies_per_cluster[cluster].append(item)
  else:
    config_policies_per_cluster[cluster] = [item]

policies_per_cluster = {}
for key in config_policies_per_cluster.keys():
  policies = {}
  config_policies = config_policies_per_cluster[key]
  for config_policy in config_policies:
    policy_id = config_policy['metadata']['labels']['policy-id']
    if policy_id in policies:
      policies[policy_id].append(config_policy)
    else:
      policies[policy_id] = [config_policy]
  policies_per_cluster[key] = policies

NOT_COMPLIANT = 'NotCompliant'
header = ["Cluster", "Policy", "Status"]
column_lengths = list(map(lambda x: len(x), header))
rows = [header]
for cluster in policies_per_cluster:
  for policy in policies_per_cluster[cluster]:
    status_summary = NOT_COMPLIANT
    config_policies = policies_per_cluster[cluster][policy]
    # for config_policy in config_policies:
    #   status = config_policy['status']
    #   if status != None:
    #     compliant = status['compliant']
    row = [cluster, policy, status_summary]
    rows.append(row)

for row in rows:
  for column_idx, column in enumerate(row):
    if len(column) > column_lengths[column_idx]:
      column_lengths[column_idx] = len(column)

formatted_rows = []
for row in rows:
  formatted_row = []
  for column_idx, column in enumerate(row):
    formatted_row.append(column.ljust(column_lengths[column_idx]))
  formatted_rows.append(formatted_row)

for formatted_row in formatted_rows:
  print("\t".join([str(x) for x in formatted_row]))