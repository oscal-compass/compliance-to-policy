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

from typing import Any, Dict, Optional

from pydantic.v1 import Field

from c2p.common.c2p_base_model import C2PBaseModel


class Metadata(C2PBaseModel):
    """
    Attributes:
        filepath: Filepath
    """

    filepath: Optional[str] = Field(None, title='Filepath')


class RawResult(C2PBaseModel):
    """

    Attributes:
        metadata: Metadata
        data: Data
        additional_props: Additional properties
    """

    metadata: Metadata = Field(Metadata())
    data: Any = Field(title='Serialized raw results (JSON, YAML) as dict object')
    additional_props: Optional[Dict[str, Any]] = Field(
        {}, title='Additional properties', description='Add any information in key-value format if required.'
    )
