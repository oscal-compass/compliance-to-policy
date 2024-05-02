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

import argparse
import pathlib

from trestle.common.err import handle_generic_command_exception
from trestle.core.commands.command_docs import CommandBase
from trestle.core.commands.common.return_codes import CmdReturnCodes

from c2p.common import logging
from c2p.tools.oscal_csv_to_json import OscalCsvToJson

logger = logging.getLogger(__name__)


class Csv2OscalCd(CommandBase):
    """Command to generate OSCAL Component Definition from component definition in csv format"""

    name = 'csv-to-oscal-cd'

    def _init_arguments(self) -> None:
        self.add_argument(
            '-c',
            '--config',
            type=pathlib.Path,
            help='Path to config file if --csv, --title, and -o are not given',
            required=False,
        )
        self.add_argument('--title', type=str, help='Title of component-definition', required=False)
        self.add_argument('--csv', type=pathlib.Path, help='Path to csv file', required=False)
        self.add_argument(
            '-o',
            '--out',
            type=pathlib.Path,
            help='Path to directory for output of component-definition.json',
            required=False,
        )
        self.add_argument(
            '-i', '--info', action='store_true', help='Print information about a particular task.', required=False
        )

    def _run(self, args: argparse.Namespace) -> int:
        octj = OscalCsvToJson()
        try:
            if args.config != None:
                octj.generate(pathlib.Path(args.config))
            elif args.title != None and args.csv != None and args.out != None:
                path = octj.generate_config(args.title, args.csv, args.out)
                octj.generate(path)

        except Exception as e:
            return handle_generic_command_exception(
                e, logger, 'Error while performing OSCAL Assessment Results generation'
            )

        return CmdReturnCodes.SUCCESS.value
