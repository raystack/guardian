# Create Your Second Policy 
In this example we will explain some more details around the policy configuartions. Guardian can connect to an external identity manager to retrieve user details information. When a user creates an appeal using the policy given below, Guardian will connect to **`http://youridentitymanager.com/api/users/{user_id}`** for taking the user information defined in the **`iam_schema`** within the policy. 

### Policy Example
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
  description: resource owner approval. Will skip this for playground dataset
  strategy: manual
  when: not ($appeal.resource.type == "dataset" && $appeal.resource.urn == "my-bq-project:playground")
  approvers:
  - $appeal.resource.details.owner
```

### Explanation
For the approval, a user's appeal will follow the steps **`employee_check`** and **`resource_owner_approval`** in the same order.
The first step is an **`auto`** strategy which checks the pre-defined condition that the employee who is requesting for the access belongs to the same company. Until then the status of the appeal will be **`pending`** for the first step(**`employee_check`**), and **`blocked`** for the second step(**`resource_owner_approval`**).

Once this is approved, the status is updated to **`approved`** and **`pending`** for the two steps respectively. The **`when`** field contains the condition for which the step can be skipped. In this case, if the appeal is for a playground dataset, the resource owner approval is not required, otherwise owner's approval is required to get the status of appeal to **`active`**.
