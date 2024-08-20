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

from abc import ABC, abstractmethod
from typing import Any

from pydantic.v1 import BaseModel

from c2p.framework.models.policy import Policy
from c2p.framework.models.pvp_result import PVPResult
from c2p.framework.models.raw_result import RawResult

PluginConfig = BaseModel


class PluginSpec(ABC):

    @abstractmethod
    def generate_pvp_result(self, raw_result: RawResult) -> PVPResult:
        pass

    @abstractmethod
    def generate_pvp_policy(self, policy: Policy) -> Any:
        pass
