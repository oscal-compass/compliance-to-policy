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

import pathlib
from typing import Dict, List, Union

import yaml


def load_yaml(path: Union[str, pathlib.Path]) -> Dict:
    if isinstance(path, str):
        path = pathlib.Path(path)
    return yaml.safe_load(path.open('r'))


def load_yamls(path: Union[str, pathlib.Path]) -> List[Dict]:
    if isinstance(path, str):
        path = pathlib.Path(path)
    return yaml.safe_load_all(path.open('r'))
