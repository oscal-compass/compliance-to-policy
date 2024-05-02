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

import os
import pathlib
from distutils.util import strtobool

from c2p.common import logging

logger = logging.getLogger(__name__)


def write_test_result_to_file(output_path: pathlib.Path, data: str):
    enabled = os.getenv('ENABLE_WRITE_TEST_RESULT_TO_FILE', 'false')
    if strtobool(enabled):
        try:
            output_path.write_text(data)
        except Exception as e:
            logger.error(f'Failed to write test results to {output_path.as_posix()}\n {str(e)}')
