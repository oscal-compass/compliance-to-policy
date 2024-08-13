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

import datetime
import json
import pathlib
from typing import Any, Dict, Optional, Type, TypeVar

import orjson
from pydantic.v1 import BaseModel
from pydantic.v1.parse import load_file
from trestle.core.base_model import robust_datetime_serialization

import c2p.common.err as err


class C2PBaseModel(BaseModel):
    """Base Model. Serves as wrapper around BaseModel for overriding methods."""

    class Config:
        json_encoders = {datetime.datetime: lambda x: robust_datetime_serialization(x)}

    @classmethod
    def read(cls, path: pathlib.Path) -> Optional['C2PBaseModel']:
        obj: Dict[str, Any] = {}
        try:
            obj = load_file(
                path,
                json_loads=cls.__config__.json_loads,
            )
        except Exception as e:
            raise err.C2PError(f'Error loading file {path} {str(e)}')

        try:
            parsed = cls.parse_obj(obj)
        except Exception as e:
            raise err.C2PError(f'Error parsing file {path} {str(e)}')

        return parsed

    def serialize_json_bytes(self, pretty: bool = False) -> bytes:
        odict = self.dict(by_alias=True, exclude_none=True)
        if pretty:
            return orjson.dumps(odict, default=self.__json_encoder__, option=orjson.OPT_INDENT_2)  # type: ignore
        return orjson.dumps(odict, default=self.__json_encoder__)  # type: ignore


T = TypeVar('T', BaseModel, Any)


class C2PBaseDict(Dict[str, T]):
    _member_class: Type[T]

    def __init__(self, obj: dict = {}, **kwargs):
        if issubclass(self._member_class, BaseModel):
            member_dict = {}
            for key, value in kwargs.items():
                c = self._member_class(**value)
                member_dict[key] = c
            for key, value in obj.items():
                c = self._member_class(**value)
                member_dict[key] = c
            super().__init__(member_dict)
        else:
            super().__init__(obj, **kwargs)

    def serialize_json_bytes(self, pretty: bool = False) -> bytes:
        if pretty:
            return orjson.dumps(self, default=self._get_json_encoder(), option=orjson.OPT_INDENT_2)  # type: ignore
        return orjson.dumps(self, default=self._get_json_encoder())

    def json(self) -> str:
        return json.dumps(self, default=self._get_json_encoder())

    def _get_json_encoder(self) -> Any:
        if issubclass(self._member_class, BaseModel):
            return self._member_class.__json_encoder__
        else:
            return None
