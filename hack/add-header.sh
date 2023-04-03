# Copyright 2023 IBM Corporation

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#!/bin/bash

header=$1
dir=$2
ext=$3

license=`cat $header`

find $dir -name "*.$ext" | while read file
do
  if grep -q "Copyright" $file; then
    echo "Already copyrighted $file"
  else
    echo "Add copy right $file"
    (cat "$header" ; echo "") | cat - "$file" > /tmp/tmp.txt
    mv /tmp/tmp.txt $file
  fi
done