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
from trestle.oscal.assessment_results import AssessmentResults
from trestle.oscal.catalog import Catalog
from trestle.oscal.component import ComponentDefinition

from c2p.common import logging
from c2p.tools.posture import posture, render
from c2p.tools.posture.model import Source

logger = logging.getLogger(__name__)


class Result2Posture(CommandBase):
    """Command to render OSCAL Assessment Results as a structured compliance posture"""

    name = 'result2posture'

    def _init_arguments(self) -> None:
        self.add_argument(
            '-ar',
            '--assessment-results',
            type=pathlib.Path,
            action='append',
            help='Path to OSCAL Assessment Results. Repeat the flag to combine several: a c2p '
            'run assesses one PVP at a time, so a cluster with several policy engines produces '
            'one document per engine and the posture is where they come back together',
            required=True,
        )
        self.add_argument(
            '-cdef',
            '--component-definition',
            type=pathlib.Path,
            help='Path to OSCAL Component Definition. The posture is accountable for every '
            'check this declares, which is what lets an unassessed control be reported as such',
            required=True,
        )
        self.add_argument(
            '-cat',
            '--catalog',
            type=pathlib.Path,
            help='Path to OSCAL Catalog, used to title the controls',
            required=False,
        )
        self.add_argument(
            '-n',
            '--framework',
            type=str,
            help='Compliance framework name to record in the posture',
            required=False,
        )
        self.add_argument(
            '-f',
            '--format',
            type=str,
            choices=['json', 'md', 'html'],
            default='json',
            help='Output format. json is the c2p-posture/v1 document; md and html are views over it',
            required=False,
        )
        self.add_argument(
            '-o',
            '--out',
            type=pathlib.Path,
            help='Path to output file. Omit to write to stdout',
            required=False,
        )

    def _run(self, args: argparse.Namespace) -> int:
        try:
            ar = [AssessmentResults.oscal_read(path) for path in args.assessment_results]
            cdef = ComponentDefinition.oscal_read(args.component_definition)
            catalog = Catalog.oscal_read(args.catalog) if args.catalog != None else None
            source = Source(
                assessment_results=', '.join(str(path) for path in args.assessment_results),
                component_definition=str(args.component_definition),
                catalog=str(args.catalog) if args.catalog != None else None,
            )
            result = posture.generate(ar, cdef, catalog=catalog, framework=args.framework, source=source)
            renderers = {'json': render.render_json, 'md': render.render_markdown, 'html': render.render_html}
            rendered = renderers[args.format](result)

            if args.out != None:
                pathlib.Path(args.out).open('w').write(rendered)
            else:
                self.out(rendered)

        except Exception as e:
            return handle_generic_command_exception(e, logger, 'Error while generating the compliance posture')

        return CmdReturnCodes.SUCCESS.value
