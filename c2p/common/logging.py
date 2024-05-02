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

import logging
import sys

# Singleton logger instance
_logger = logging.getLogger('c2p')

FORMATTER_STR = '[%(asctime)s %(levelname)s %(name)s] %(message)s'


def set_global_logging_levels(level: int = logging.INFO) -> None:
    """Initialise logging.

    Should only be invoked by the CLI classes or similar.
    """
    # This line stops default root loggers setup for a python context from logging extra messages.
    # DO NOT USE THIS COMMAND directly from an SDK. Handle logs levels based on your own application
    _logger.propagate = False
    # Remove handlers
    _logger.handlers.clear()
    # set global level
    _logger.setLevel(level)
    # Create standard out
    handler = logging.StreamHandler(sys.stderr)
    handler.setLevel(level)
    handler.setFormatter(logging.Formatter(FORMATTER_STR))
    # add ch to logger
    _logger.addHandler(handler)


def getLogger(name: str) -> logging.Logger:
    return logging.getLogger(name)
