# Jobs

## Server Jobs Configurations

```yaml
JOBS:
  FETCH_RESOURCES:
    ENABLED: true
    INTERVAL: '0 */2 * * *'  #"At minute 0 past every 2nd hour"
  REVOKE_EXPIRED_ACCESS:
    ENABLED: true
    INTERVAL: '*/20 * * * *'  #“At every 20th minute"
  EXPIRING_ACCESS_NOTIFICATION:
    ENABLED: true
    INTERVAL: '0 9 * * *' #"At minute 0 past hour 9"
```

| Field   | Description                   | 
| ------- | ----------------------------- | 
| `FETCH_RESOURCES`    | When Enabled, the Guardian server fetches resources from the providers and updated the database.        | 
| `REVOKE_EXPIRED_ACCESS` | When Enabled, the Guardian server will revoke the user permissions for the resource |
| `EXPIRING_ACCESS_NOTIFICATION`   | When Enabled, the Guardian server will notify the user on the notifier (currently `slack` only) before the user appeal is about to expire.<br/><br/>The user gets notified before 7 days, 3 days and 1 day of appeal expiry    | 
