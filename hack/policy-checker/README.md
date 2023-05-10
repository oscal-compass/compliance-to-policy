## How to run
1. Install modules
```
pip install -r requirements
```
1. Run
```
$ python check.py --policy-collection-dir ../../policy-collection/community --regenerated-policy-dir ./work/test1-tmp/generated/nist-high-eu/deliverable-policies

+-------------------------------------+--------+
|                                     | Result |
+-------------------------------------+--------+
|       # of generated policies       |   89   |
|  # of consistent with original one  |   85   |
| # of inconsistent with original one |   4    |
+-------------------------------------+--------+
List of inconsistent policie
['policy-trusted-node', 'policy-trusted-container', 'policy-ocs-machinesets', 'policy-cert-ocp4']
```