# Create Your First Appeal 

**Note:**
1. Get the `resource_id` of Playground dataset in our example ([Steps](./update-resource#getting-the-resourceid-for-the-resource))
2. Currently we support creating an Appeal via the API only

** Here is an example below: **

```bash
$ curl --request POST '{{HOST}}/api/v1beta1/appeals' \
--header 'X-Auth-Email: user@company.com' \
--header 'Content-Type: application/json' \
--data-raw '{
  "account_id": "user@company.com",
  "resources": [
    {
      "id": "<<playground resource id>>",
      "role": "viewer"
    }
  ]
}'
```
**Note:** Refer to the [Appeal Request](../reference/api#appeal-request-config) Configurations for more details

**The Response after creating the appeal is as follows:**

```json
{
    "appeals": [
        {
            "id": "<< appeal id >>",
            "resource_id": "<< playground resource id >>",
            "policy_id": "my-first-policy",
            "policy_version": 1,
            "status": "pending",
            "account_id": "user@company.com",
            "role": "viewer",
            "resource": {
                "id": "<< playground resource id >>",
                "provider_type": "bigquery",
                "provider_urn": "my-first-bigquery-provider",
                "type": "dataset",
                "urn": "<<my-bq-project-id>>:playground",
                "name": "playground",
                "details": {
                    "owner": "owner.guy@company.com"
                },
                "created_at": "2022-06-30T10:46:03.608245Z",
                "updated_at": "2022-06-30T10:50:22.966110Z"
            },
            "approvals": [
                {
                    "id": "<< approval id 1 >>",
                    "name": "resource_owner_approval",
                    "appeal_id": "<< appeal id >>",
                    "status": "pending",
                    "policy_id": "my-first-policy",
                    "policy_version": 1,
                    "approvers": [
                        "owner.guy@company.com"
                    ],
                    "created_at": "2022-06-30T10:55:48.712177Z",
                    "updated_at": "2022-06-30T10:55:48.712177Z"
                },
                {
                    "id": "<< approval id 2 >>",
                    "name": "admin_approval",
                    "appeal_id": "<< appeal id >>",
                    "status": "blocked",
                    "policy_id": "my-first-policy",
                    "policy_version": 1,
                    "approvers": [
                        "john.doe@company.com"
                    ],
                    "created_at": "2022-06-30T10:55:48.712177Z",
                    "updated_at": "2022-06-30T10:55:48.712177Z"
                }
            ],
            "created_at": "2022-06-30T10:55:48.704006Z",
            "updated_at": "2022-06-30T10:55:48.704006Z",
            "revoked_at": "0001-01-01T00:00:00Z",
            "details": {},
            "account_type": "user",
            "created_by": "user@company.com",
            "creator": null
        }
    ]
}
```