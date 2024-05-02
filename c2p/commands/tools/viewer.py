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
from trestle.oscal.assessment_results import AssessmentResults
from trestle.oscal.component import ComponentDefinition

from c2p.common import logging
from c2p.tools.viewer import viewer

logger = logging.getLogger(__name__)


class Viewer(CommandBase):
    """Command to render OSCAL Assessment Results in markdown"""

    name = 'viewer'

    def _init_arguments(self) -> None:
        self.add_argument(
            '-ar',
            '--assessment-results',
            type=pathlib.Path,
            help='Path to OSCAL Assessment Results',
            required=True,
        )
        self.add_argument(
            '-cdef',
            '--component-definition',
            type=pathlib.Path,
            help='Path to OSCAL Component Definition',
            required=True,
        )
        self.add_argument(
            '-o',
            '--out',
            type=pathlib.Path,
            help='Path to output file',
            required=False,
        )

    def _run(self, args: argparse.Namespace) -> int:

        ar = AssessmentResults.oscal_read(args.assessment_results)
        cdef = ComponentDefinition.oscal_read(args.component_definition)
        rendered_md = viewer.render(ar, cdef)
        try:
            if args.out != None:
                pathlib.Path(args.out).open('w').write(rendered_md)
            else:
                self.out(rendered_md)

        except Exception as e:
            return handle_generic_command_exception(e, logger, 'Error while performing rendering Assessment Results')

        return CmdReturnCodes.SUCCESS.value
