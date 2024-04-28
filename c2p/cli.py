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
from logging import DEBUG
from sys import exit

from trestle.common import const, log
from trestle.core.commands.command_docs import CommandBase

from c2p.commands.tools.tools import Tools
from c2p.commands.version import VersionCmd
from c2p.common import logging

logger = logging.getLogger(__name__)


class C2P(CommandBase):
    """Bridge Compliance and Policy"""

    subcommands = [
        VersionCmd,
        Tools,
    ]

    def _init_arguments(self) -> None:
        self.add_argument('-v', '--verbose', help=const.DISPLAY_VERBOSE_OUTPUT, action='count', default=0)

    def _validate_and_run(self, args: argparse.ArgumentParser):
        if args.verbose > 0:
            logging.set_global_logging_levels(DEBUG)


def run() -> None:
    """Run the c2p cli."""
    log.set_global_logging_levels()
    logging.set_global_logging_levels()
    logger.debug('Main entry point.')

    exit(C2P().run())
