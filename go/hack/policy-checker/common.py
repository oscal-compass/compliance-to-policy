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
from yaml import Loader
import os

def parse(manifest, filename):
  if manifest['kind'] != 'Policy':
    return None
  config_policies = []
  metadata = manifest['metadata']
  if 'annotations' in metadata:
    annotations = metadata['annotations']
    standards = annotations['policy.open-cluster-management.io/standards']
    categories = annotations['policy.open-cluster-management.io/categories']
    controls = annotations['policy.open-cluster-management.io/controls']
  else:
    standards = ''
    categories = ''
    controls = ''
  spec = manifest['spec']
  policy_templates = spec['policy-templates']
  for policy_template in policy_templates:
    object_definition = policy_template['objectDefinition']
    kind = object_definition['kind']
    if kind == 'ConfigurationPolicy':
      object_templates = object_definition['spec']['object-templates']
      resources = []
      for object_template in object_templates:
        config_object_definition = object_template['objectDefinition']
        resource = {
          'kind': config_object_definition['kind'],
          'apiVersion': config_object_definition['apiVersion'],
          'name': config_object_definition['metadata']['name'] if 'metadata' in config_object_definition and 'name' in config_object_definition['metadata'] else '',
        }
        resources.append(resource)
      config_policy = {
        'name': object_definition['metadata']['name'],
        'resources': resources
      }
      config_policies.append(config_policy)
  return {
    'filename': filename,
    'name': metadata['name'],
    'standards': parse_compliance_annotation(standards),
    'categories': parse_compliance_annotation(categories),
    'controls': parse_compliance_annotation(controls),
    'config_policies': config_policies,
  }

def parse_compliance_annotation(ann):
  anns = ann.split(',')
  newone = []
  for x in anns:
    newone.append(x.strip())
  return newone

def load(path):
  manifests = []
  with open(path, 'r') as file:
      docs = yaml.safe_load_all(file)
      for doc in docs:
          # pprint.pprint(doc)
          manifests.append(doc)
  return manifests

def walkthrough(files):
  policies = []
  for file in files:
    filename = os.path.basename(file)
    manifests = load(file)
    for manifest in manifests:
      if manifest == None or not 'kind' in manifest:
        continue
      policy = parse(manifest, filename)
      if policy != None:
        policies.append(policy)
  return policies