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
"""Make the reference implementation importable from its own tests.

The directory above this one has a hyphen in its path, so the package beside these tests cannot
be imported as a dotted module and is put on the path here instead.

The reference implementation deploys policy to a cluster and reads engine messages back into a
document an auditor may read. That is precisely the code that should not be the only thing in
the repository without a test, even though it is an example rather than library code.
"""

import pathlib
import sys

REFERENCE = pathlib.Path(__file__).resolve().parents[1]
if str(REFERENCE) not in sys.path:
    sys.path.insert(0, str(REFERENCE))
