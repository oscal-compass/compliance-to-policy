# Assessment Results from Prowler

Produced by:

```bash
python samples_public/prowler/compliance_to_policy.py -o /tmp/prowler-policy
python samples_public/prowler/result_to_compliance.py -d /tmp/prowler-policy
```

The sample data is chosen so that one run exercises every outcome the plugin can
produce: a clean pass, a partial failure, a muted finding and a manual check.

| Rule | State | Reason |
| --- | --- | --- |
| `aws-root-mfa` | satisfied | `pass` |
| `aws-console-mfa` | not-satisfied | `fail` — one console user without MFA |
| `aws-admin-mfa` | not-satisfied | `other` — muted, so suppressed rather than proven |
| `aws-audit-trail` | not-satisfied | `other` — a manual check, awaiting determination |

```json
{
  "assessment-results": {
    "uuid": "48025a0f-34e0-4b9b-895a-81b5d01ac77d",
    "metadata": {
      "title": "Prowler Assessment Results",
      "last-modified": "2026-08-11T21:22:56+00:00",
      "version": "3.12.4",
      "oscal-version": "1.1.3"
    },
    "import-ap": {
      "href": "https://not-available-for-now"
    },
    "results": [
      {
        "uuid": "f6df859d-ffcb-4709-b566-854b5babb5c0",
        "title": "Prowler Assessment Results",
        "description": "OSCAL Assessment Results from Prowler",
        "start": "2026-08-11T21:22:56+00:00",
        "reviewed-controls": {
          "control-selections": [
            {
              "include-controls": [
                {
                  "control-id": "ia-2.1",
                  "statement-ids": []
                },
                {
                  "control-id": "ia-2",
                  "statement-ids": []
                },
                {
                  "control-id": "ac-6.2",
                  "statement-ids": []
                },
                {
                  "control-id": "au-2",
                  "statement-ids": []
                }
              ]
            }
          ]
        },
        "observations": [
          {
            "uuid": "26e95903-467d-4ab1-9a51-0919690e3a7b",
            "title": "cloudtrail_multi_region_enabled",
            "description": "cloudtrail_multi_region_enabled",
            "props": [
              {
                "name": "assessment-rule-id",
                "value": "aws-audit-trail"
              },
              {
                "name": "assessment-check-id",
                "value": "cloudtrail_multi_region_enabled"
              },
              {
                "name": "resources-evaluated",
                "value": "1"
              }
            ],
            "methods": [
              "AUTOMATED"
            ],
            "subjects": [
              {
                "subject-uuid": "a8f9eb58-56a8-4858-b904-cc8bdf053dfb",
                "type": "resource",
                "title": "AwsCloudTrailTrail main us-east-1",
                "props": [
                  {
                    "name": "resource-id",
                    "value": "arn:aws:cloudtrail:us-east-1:123456789012:trail/main"
                  },
                  {
                    "name": "result",
                    "value": "error"
                  },
                  {
                    "name": "evaluated-on",
                    "value": "2026-08-11T21:22:56+00:00"
                  },
                  {
                    "name": "reason",
                    "value": "cloudtrail_multi_region_enabled: MANUAL"
                  }
                ]
              }
            ],
            "collected": "2026-08-11T21:22:56+00:00"
          },
          {
            "uuid": "eb855f23-adc7-44fa-b296-2fa5e8c6404b",
            "title": "iam_administrator_access_with_mfa",
            "description": "iam_administrator_access_with_mfa",
            "props": [
              {
                "name": "assessment-rule-id",
                "value": "aws-admin-mfa"
              },
              {
                "name": "assessment-check-id",
                "value": "iam_administrator_access_with_mfa"
              },
              {
                "name": "resources-evaluated",
                "value": "1"
              },
              {
                "name": "resources-muted",
                "value": "1"
              }
            ],
            "methods": [
              "AUTOMATED"
            ],
            "subjects": [
              {
                "subject-uuid": "218c39a1-eb46-4c95-baed-cb545002cb0b",
                "type": "resource",
                "title": "AwsIamGroup admins us-east-1",
                "props": [
                  {
                    "name": "resource-id",
                    "value": "arn:aws:iam::123456789012:group/admins"
                  },
                  {
                    "name": "result",
                    "value": "error"
                  },
                  {
                    "name": "evaluated-on",
                    "value": "2026-08-11T21:22:56+00:00"
                  },
                  {
                    "name": "reason",
                    "value": "iam_administrator_access_with_mfa: MUTED"
                  }
                ]
              }
            ],
            "collected": "2026-08-11T21:22:56+00:00"
          },
          {
            "uuid": "76dd5532-81da-4fd1-92e9-8f4dd79902ac",
            "title": "iam_root_mfa_enabled",
            "description": "iam_root_mfa_enabled",
            "props": [
              {
                "name": "assessment-rule-id",
                "value": "aws-root-mfa"
              },
              {
                "name": "assessment-check-id",
                "value": "iam_root_mfa_enabled"
              },
              {
                "name": "resources-evaluated",
                "value": "1"
              }
            ],
            "methods": [
              "AUTOMATED"
            ],
            "subjects": [
              {
                "subject-uuid": "17f64fe9-72cf-4a1b-a957-5191246998ea",
                "type": "resource",
                "title": "AwsIamUser root us-east-1",
                "props": [
                  {
                    "name": "resource-id",
                    "value": "arn:aws:iam::123456789012:root"
                  },
                  {
                    "name": "result",
                    "value": "pass"
                  },
                  {
                    "name": "evaluated-on",
                    "value": "2026-08-11T21:22:56+00:00"
                  },
                  {
                    "name": "reason",
                    "value": "iam_root_mfa_enabled: PASS"
                  }
                ]
              }
            ],
            "collected": "2026-08-11T21:22:56+00:00"
          },
          {
            "uuid": "8bd8f91e-3115-4183-a224-1eba9c897b64",
            "title": "iam_user_mfa_enabled_console_access",
            "description": "iam_user_mfa_enabled_console_access",
            "props": [
              {
                "name": "assessment-rule-id",
                "value": "aws-console-mfa"
              },
              {
                "name": "assessment-check-id",
                "value": "iam_user_mfa_enabled_console_access"
              },
              {
                "name": "resources-evaluated",
                "value": "2"
              }
            ],
            "methods": [
              "AUTOMATED"
            ],
            "subjects": [
              {
                "subject-uuid": "c09fb542-a3c3-42fa-8d3c-05928fa22ff1",
                "type": "resource",
                "title": "AwsIamUser alice us-east-1",
                "props": [
                  {
                    "name": "resource-id",
                    "value": "arn:aws:iam::123456789012:user/alice"
                  },
                  {
                    "name": "result",
                    "value": "pass"
                  },
                  {
                    "name": "evaluated-on",
                    "value": "2026-08-11T21:22:56+00:00"
                  },
                  {
                    "name": "reason",
                    "value": "iam_user_mfa_enabled_console_access: PASS"
                  }
                ]
              },
              {
                "subject-uuid": "c2c783b4-9a25-469e-83da-8020d2cd62f8",
                "type": "resource",
                "title": "AwsIamUser contractor us-east-1",
                "props": [
                  {
                    "name": "resource-id",
                    "value": "arn:aws:iam::123456789012:user/contractor"
                  },
                  {
                    "name": "result",
                    "value": "failure"
                  },
                  {
                    "name": "evaluated-on",
                    "value": "2026-08-11T21:22:56+00:00"
                  },
                  {
                    "name": "reason",
                    "value": "iam_user_mfa_enabled_console_access: FAIL"
                  }
                ]
              }
            ],
            "collected": "2026-08-11T21:22:56+00:00"
          }
        ],
        "findings": [
          {
            "uuid": "4561ac7c-c50a-468f-abcf-5e978da29dee",
            "title": "aws-admin-mfa",
            "description": "Not proven: iam_administrator_access_with_mfa (error or suppressed)",
            "target": {
              "type": "objective-id",
              "target-id": "aws-admin-mfa",
              "status": {
                "state": "not-satisfied",
                "reason": "other",
                "remarks": "Not proven: iam_administrator_access_with_mfa (error or suppressed)"
              }
            }
          },
          {
            "uuid": "1b7d15af-96bc-4272-9f56-46e5b9b19b0c",
            "title": "aws-audit-trail",
            "description": "Not proven: cloudtrail_multi_region_enabled (error or suppressed)",
            "target": {
              "type": "objective-id",
              "target-id": "aws-audit-trail",
              "status": {
                "state": "not-satisfied",
                "reason": "other",
                "remarks": "Not proven: cloudtrail_multi_region_enabled (error or suppressed)"
              }
            }
          },
          {
            "uuid": "9a7994a7-d2f9-416f-8d1c-203498f01dec",
            "title": "aws-console-mfa",
            "description": "Failing checks: iam_user_mfa_enabled_console_access",
            "target": {
              "type": "objective-id",
              "target-id": "aws-console-mfa",
              "status": {
                "state": "not-satisfied",
                "reason": "fail",
                "remarks": "Failing checks: iam_user_mfa_enabled_console_access"
              }
            }
          },
          {
            "uuid": "bc36512f-4b84-4af5-a7c1-8a8ab0a6eecc",
            "title": "aws-root-mfa",
            "description": "All checks passed on every evaluated resource: iam_root_mfa_enabled",
            "target": {
              "type": "objective-id",
              "target-id": "aws-root-mfa",
              "status": {
                "state": "satisfied",
                "reason": "pass",
                "remarks": "All checks passed on every evaluated resource: iam_root_mfa_enabled"
              }
            }
          }
        ]
      }
    ]
  }
}
```
