# Roadmap

In the following section, you can learn about what features we're working on, what stage they're in, and when we expect to bring them to you. Have any questions or comments about items on the roadmap? Join the [discussions](https://github.com/orgs/odpf/discussions) on the Guardian Github forum.

We’re planning to iterate on the format of the roadmap itself, and we see the potential to engage more in discussions about the future of Firehose features. If you have feedback about this roadmap section itself, such as how the issues are presented, let us know through [discussions](https://github.com/orgs/odpf/discussions).

Guardian roadmap can be tracked on this [project](https://github.com/orgs/odpf/projects/10/views/4). The roadmap is arranged on a project board to give a sense for how far out each item is on the horizon. Every product or feature is added to a particular project board column according to the quarter in which it is expected to ship next.

Here, we outline some (but not all!) of the products on our roadmap. We'd love your input and feedback, Contact us to discuss any of the below, or any other products you'd like to see.

## Appeal approval outside Guardian

Complex appeal approval flow can't be modelled using policy config YAML file so in that case, Guardian should be able to integrate with existing complex approval flow like `bpmn`.

Proposed solution is Guardian can integrate with exiting approval flow either by webhook or subscribe to events.

## Role based auto appeal to multiple resources

Provide a user to have pre-defined access based on their role. For example, if a developer/analyst joins a team then they will have access of certain Metabase Collection, BigQuery dataset etc.

There are two ways to approach this:
- Create a role-resource mapping table, Api to crud operation on this table. Cronjob or a request flow to trigger default appeals based on role and role-resource mapping table.
- Able to tag resources and create a role-tag table. Cronjob or a request flow to trigger default appeals based on role and role-tag mapping table.

## Data access monitoring

Guardian can manage access across multiple providers. But it is still hard for data governance managers to monitor the different aspects of data access.

**Goal**:<br/>
With data access monitoring in Guardian, we aim to provide answers to the following questions.
- How many users have access to sensitive data?
- How many appeals are pending?
- How many appeals are about to expire?
- What kind of data authorized users are accessing?
- When was a resource accessed, by whom, and for what purpose?
- Answers to these questions are very important to be proactive in managing security and compliance.

**Scope**:<br/>
Access monitoring can be tracked across different sections
- Appeals: Analytics about appeals and their status.
- Access Logs: Analytics about what resources are being actually queried and how frequently.
- Users: Analytics about how many users are active, and have access across resources.
- Resources: Analytics about how many resources are available and of what type.

## Add support for Postgres provider

We need to support access management of Postgres. The proposed solution:
- Provider configuration for Postgres
- Postgres client
- Postgres resource & access management (TODO: figure out what resources that need to be granted & revoked)
- Documentation

## CLI authentication

some APIs expect an authorized user email to be present in the request header (metadata for GRPC). For example approve/reject approval API. On web, rely on the auth proxy to provide this header. But for guardian CLI we haven't had an auth mechanism for accessing those APIs.

Need to provide configurable authentication method in the CLI