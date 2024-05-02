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

import configparser
import pathlib
from textwrap import dedent
from typing import Optional

import trestle.common.const as const
from trestle.common.err import TrestleError
from trestle.tasks.base_task import TaskOutcome
from trestle.tasks.csv_to_oscal_cd import CsvToOscalComponentDefinition

from c2p.common.logging import getLogger

logger = getLogger(__name__)


class OscalCsvToJson:
    def __init__(self) -> None:
        pass

    def generate_config(self, title: str, csv_path: pathlib.Path, output_path: pathlib.Path) -> pathlib.Path:
        path = output_path / 'csv-to-oscal-cd.config'
        with open(path.as_posix(), 'w') as file:
            data = f"""
            [task.csv-to-oscal-cd]

            title = {title}
            version = 1.0
            csv-file = {csv_path.as_posix()}
            output-dir = {output_path.as_posix()}
            """
            file.write(dedent(data))
        return path

    def generate(self, config_path: pathlib.Path):
        config = configparser.ConfigParser(interpolation=configparser.ExtendedInterpolation())
        config.read_file(config_path.open('r', encoding=const.FILE_ENCODING))
        config_section: Optional[configparser.SectionProxy] = None
        section_label = 'task.csv-to-oscal-cd'
        if section_label in config.sections():
            config_section = config[section_label]
        else:
            logger.warning(
                f'Config file was not configured with the appropriate section for the task: "[{section_label}]"'
            )
        task = CsvToOscalComponentDefinition(config_section)
        simulate_result = task.simulate()
        if not (simulate_result == TaskOutcome.SIM_SUCCESS):
            raise TrestleError(f'Task {section_label} reported a {simulate_result}')

        actual_result = task.execute()
        if not (actual_result == TaskOutcome.SUCCESS):
            raise TrestleError(f'Task {section_label} reported a {actual_result}')

        logger.info(f'Task: {section_label} executed successfully.')
