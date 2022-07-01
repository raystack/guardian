

## Complex example
```yaml
id: my-second-policy
iam:
  provider: http
  config:
    url: http://youridentitymanager.com/api/users/{user_id}
  schema:
    email: email
    name: fullName
    company: companyName
steps:
- name: employee_check
  description: only allow employee to access our resources
  strategy: auto
  approve_if: $appeal.creator.company == "Company Name"
- name: resource_owner_approval
  description: resource owner approval. will skip this for playground dataset
  strategy: manual
  when: not ($appeal.resource.type == "dataset" && $appeal.resource.urn == "my-bq-project:playground")
  approvers:
  - $appeal.resource.details.owner
```

## iam
guardian can connect to an external identity manager to retrieve user details information. in this case, when someone creating an appeal, guardian will connect to `http://youridentitymanager.com/api/users/{user_id}`

steps

first step is to check if the user is employee or not. it will auto 