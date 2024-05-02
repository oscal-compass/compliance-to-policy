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

import glob, yaml
import common
import argparse

parser = argparse.ArgumentParser()
parser.add_argument('--regenerated-policy-dir')

args = parser.parse_args() 
regenerated_policy_dir = args.regenerated_policy_dir


files = glob.glob(regenerated_policy_dir + '/*.yml', recursive=True)
policies = common.walkthrough(files)
policiesByControl = {}
for policy in policies:
  control = policy['controls'][0]
  if not control in policiesByControl:
    policiesByControl[control] = [policy['name']]
  else:
    policiesByControl[control].append(policy['name'])

print(yaml.dump(policiesByControl))



