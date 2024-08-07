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

"""Tests for cli module."""

import sys

import pytest
from _pytest.monkeypatch import MonkeyPatch

from c2p import cli
from c2p.common import logging

logger = logging.getLogger(__name__)


def test_run(monkeypatch: MonkeyPatch) -> None:
    """Test cli call."""
    testargs = ['c2p']
    monkeypatch.setattr(sys, 'argv', testargs)
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        cli.run()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code > 0


def test_version(monkeypatch: MonkeyPatch) -> None:
    """Test cli call."""
    testargs = ['c2p', 'version']
    monkeypatch.setattr(sys, 'argv', testargs)
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        cli.run()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == 0
