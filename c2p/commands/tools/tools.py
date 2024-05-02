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

from trestle.core.commands.command_docs import CommandBase

from c2p.commands.tools.csv2oscal_cd import Csv2OscalCd
from c2p.commands.tools.viewer import Viewer


class Tools(CommandBase):
    """Subcommand for tools"""

    name = 'tools'

    subcommands = [Csv2OscalCd, Viewer]
