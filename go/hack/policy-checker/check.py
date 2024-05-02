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

import pandas as pd
from prettytable import PrettyTable
import glob, json, re
import argparse
import common

parser = argparse.ArgumentParser()
parser.add_argument('--policy-collection-dir')
parser.add_argument('--regenerated-policy-dir')

args = parser.parse_args() 
policy_collection_dir = args.policy_collection_dir
regenerated_policy_dir = args.regenerated_policy_dir

# Policy Collection
files = glob.glob(policy_collection_dir + '/**/*.yaml', recursive=True)
policies = common.walkthrough(files)
with open('policies.json', 'w') as outfile:
    json.dump(policies, outfile, indent=2)

# Regenerated policies
files = glob.glob(regenerated_policy_dir + '/*.yaml', recursive=True)
regenerated_policies = common.walkthrough(files)
with open('policies-regenerated.json', 'w') as outfile:
    json.dump(regenerated_policies, outfile, indent=2)

def find_regenerated_policy(policy, regenerated_policies):
  for regen_policy in regenerated_policies:
    if regen_policy['name'] == policy['name']:
      return regen_policy
  return None

def find_resource(resource, regen_config_policies):
  for regen_config_policy in regen_config_policies:
    for regen_resource in regen_config_policy['resources']:
      if regen_resource['kind'] == resource['kind'] and regen_resource['name'] == resource['name']:
        return True
  return False

def compare_standard(standards, regen_standards):
  return 'NIST SP 800-53' in standards

def compare_category(categories, regen_categories):
  pattern = '([A-Za-z0-9_]+) .+'
  repatter = re.compile(pattern)
  for x in categories:
    result = repatter.match(x)
    if result.lastindex > 0 and result.group(1).lower() == regen_categories[0]:
      return True
  return False

def compare_control(controls, regen_controls):
  pattern = '([A-Za-z0-9_-]+)( .+|$)'
  repatter = re.compile(pattern)
  for x in controls:
    result = repatter.match(x)
    if result != None and result.lastindex > 0 and result.group(1).lower() == regen_controls[0]:
      return True
  return False

# Compare
results = []
for policy in list(filter(lambda x: 'NIST SP 800-53' in x['standards'], policies)):
  found = False
  check_standard = False
  check_category = False
  check_control = False
  check_resources = {}
  check_resources_all = False
  regen_policy = find_regenerated_policy(policy, regenerated_policies)
  if regen_policy != None:
    found = True
    check_standard = compare_standard(policy['standards'], regen_policy['standards'])
    check_category = compare_category(policy['categories'], regen_policy['categories'])
    check_control = compare_control(policy['controls'], regen_policy['controls'])
    for config_policy in policy['config_policies']:
      for resource in config_policy['resources']:
        check_resources['{0}/{1}'.format(config_policy['name'], resource['kind'])] = find_resource(resource, regen_policy['config_policies'])
    for key in check_resources:
      check_resources_all = check_resources[key]
  
  result = {
    'name': policy['name'],
    'found': found,
    'standards': policy['standards'],
    're_standards': regen_policy['standards'] if regen_policy != None else [],
    'categories': policy['categories'],
    're_categories': regen_policy['categories']if regen_policy != None else [],
    'controls': policy['controls'],
    're_controls': regen_policy['controls']if regen_policy != None else [],
    'check_standard': check_standard,
    'check_category': check_category,
    'check_control': check_control,
    'check_resources': check_resources,
    'check_resources_all': check_resources_all,
  }
  results.append(result)

df = pd.DataFrame.from_dict(results, dtype=None)
with open('result.txt', 'w') as outfile:
  df.to_string(outfile)
  outfile.write('\n')
  outfile.write('------------------')
  outfile.write('')

df.to_csv('result.csv')

summary = []
for result in results:
  check_all = result['check_standard'] and result['check_category'] and result['check_control'] and result['check_resources_all']
  summary.append({
    'policy': result['name'],
    'is_regenerated': check_all
  })

regenerated_policies = list(filter(lambda x: x['is_regenerated'] == True, summary))
not_regenerated_policies = list(filter(lambda x: x['is_regenerated'] == False, summary))

table = PrettyTable()
table.field_names = ["", "Result"]
table.add_row(["# of generated policies", len(summary)])
table.add_row(["# of consistent with original one", len(regenerated_policies)])
table.add_row(["# of inconsistent with original one", len(not_regenerated_policies)])
print(table)

print("List of inconsistent policie")
print(list(map(lambda x: x['policy'], not_regenerated_policies)))