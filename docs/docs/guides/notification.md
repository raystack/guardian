# Notification configuration

Templates of slack notifications sent through guardian can be configured.
It is a Json string having list of blocks(Json). Developers can configure list of blocks according to the notification UI needed.
It can be list of Texts, Sections, Buttons, inputs etc (Ref:https://api.slack.com/reference/block-kit/block-elements)


## Examples:

### Only Text:

```json
[{
  "type": "section",
  "text": {
    "type": "mrkdwn", 
    "text": "You have an appeal created by {{.requestor}} requesting access to {{.resource_name}} with role {{.role}}. Appeal ID: {{.appeal_id}}"
  } 
}]
```


### Others (Sample approval notification):

```json
[
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "You have an appeal created by {{.requestor}} requesting access to {{.resource_name}} with role {{.role}}. Appeal ID: {{.appeal_id}}"
    }
  },
  {
    "type": "section",
    "fields": [
      {
        "type": "mrkdwn",
        "text": "*Provider*\\n{{.provider_type}}"
      },
      {
        "type": "mrkdwn",
        "text": "*Resource Type:*\\n{{.resource_type}}"
      }
    ]
  },
  {
    "type": "section",
    "fields": [
      {
        "type": "mrkdwn",
        "text": "*Resource:*\\n{{.resource_name}}"
      },
      {
        "type": "mrkdwn",
        "text": "*Account Id:*\\n{{.account_id}}"
      }
    ]
  },
  {
    "type": "section",
    "fields": [
      {
        "type": "mrkdwn",
        "text": "*Role:*\\n{{.role}}"
      },
      {
        "type": "mrkdwn",
        "text": "*When:*\\n{{.created_at}}"
      }
    ]
  },
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "*Console link:*\nhttps://console.io/requests/{{.appeal_id}}"
    }
  },
  {
    "type": "input",
    "element": {
      "type": "plain_text_input",
      "placeholder": {
        "type": "plain_text",
        "text": "Approve/Reject reason? (optional)"
      },
      "action_id": "reason"
    },
    "label": {
      "type": "plain_text",
      "text": "Reason"
    }
  },
  {
    "type": "actions",
    "elements": [
      {
        "text": {
          "type": "plain_text",
          "emoji": true,
          "text": "Approve"
        },
        "type": "button",
        "value": "approved",
        "style": "primary",
        "url": "https://console.io/appeal_action?action=approve&appeal_id={{.appeal_id}}&approval_step={{.approval_step}}&actor={{.actor}}"
      },
      {
        "text": {
          "type": "plain_text",
          "emoji": true,
          "text": "Reject"
        },
        "type": "button",
        "value": "rejected",
        "style": "primary",
        "url": "https://console.io/appeal_action?action=reject&appeal_id={{.appeal_id}}&approval_step={{.approval_step}}&actor={{.actor}}"
      }
    ]
  }
]
```

