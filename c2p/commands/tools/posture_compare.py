# -*- mode:python; coding:utf-8 -*-

# Copyright 2026 IBM Corporation
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
from c2p.tools.posture import compare
from c2p.tools.posture.model import Posture

logger = logging.getLogger(__name__)


class PostureCompare(CommandBase):
    """Command to compare two compliance postures control by control"""

    name = 'posture-compare'

    def _init_arguments(self) -> None:
        self.add_argument('-b', '--baseline', type=pathlib.Path, help='Path to the baseline posture', required=True)
        self.add_argument('-p', '--posture', type=pathlib.Path, help='Path to the posture to compare', required=True)
        self.add_argument('--baseline-label', type=str, default='baseline', required=False)
        self.add_argument('--posture-label', type=str, default='compared', required=False)
        self.add_argument(
            '--condition',
            type=str,
            default='in the compared posture',
            help='Completes the sentence "N control(s) stop being reported ..."',
            required=False,
        )
        self.add_argument('-f', '--format', type=str, choices=['text', 'json'], default='text', required=False)
        self.add_argument('-o', '--out', type=pathlib.Path, help='Path to output file', required=False)

    def _run(self, args: argparse.Namespace) -> int:
        try:
            baseline = Posture.read(args.baseline)
            other = Posture.read(args.posture)
            comparison = compare.compare(baseline, other)
            if args.format == 'json':
                rendered = comparison.serialize_json_bytes(pretty=True).decode('utf-8')
            else:
                rendered = compare.render_text(
                    comparison,
                    baseline_label=args.baseline_label,
                    other_label=args.posture_label,
                    condition=args.condition,
                )

            if args.out != None:
                pathlib.Path(args.out).open('w').write(rendered)
            else:
                self.out(rendered)

        except Exception as e:
            return handle_generic_command_exception(e, logger, 'Error while comparing the compliance postures')

        return CmdReturnCodes.SUCCESS.value
