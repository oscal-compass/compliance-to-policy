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
"""Kubernetes as a target platform: apply generated policy, collect what the engines saw.

**This is a reference implementation and not part of the C2P library.** It is the half of the
round trip C2P deliberately does not ship: it holds cluster credentials, it writes to a
cluster, and it reads engine messages back into a document that an auditor may read. Those are
decisions about a particular cluster, so they belong to whoever runs it. C2P's own job stops at
turning an authored control into policy, and resuming at turning collected results into OSCAL.

Read `../README.md` before running any of it.

Credentials live here and only here: a kubeconfig and a context. Nothing in the OSCAL path, and
nothing in the posture path, needs them.
"""

from .client import PROBE_RUN_LABEL, KubeClient
