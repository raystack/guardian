# Configuration
## Client Configuration

### Initialization
Client configuration can be initialized using Guardian.
To do so, use command
```bash
$ guardian config init
``` 
A `guardian.yml` file will be created in the present working directory ( with a relative path `.config/odpf/guardian.yml` ). Open this file to initialize the host as in the example here: 

```yaml
host: "localhost:8080"
```
---
## Server Configuration

### 1. Using --config file
#### Pre-requisites
- Git
- PostGres
- Golang

#### Initialization
Create a server.yml file `touch server.yml` and `open server.yml` in the pwd. Setup up a database in PostGres and provide the details in the DB field as given in the example below. For the purpose of this tutorial, we'll assume that the username is "your_user", database name is "guardian", host and port are "localhost" and 5432.

If you're new to YAML and want to learn more, see [Learn YAML in Y minutes.](https://learnxinyminutes.com/docs/yaml/)

Following is a sample job specification:

```yaml
PORT: 8080
# logging configuration
LOG:
  # debug, info, warning, error, fatal - default 'info'
  LEVEL: info
DB:                     
#For details see the Properties->Connections property of the server in PGAdmin tool
  HOST: localhost      
  USER: your_user
  PASSWORD: your_password
  NAME: guardian
  PORT: 5432          
NOTIFIER:
  PROVIDER: slack
  ACCESS_TOKEN: 
  MESSAGES:
    APPROVER_NOTIFICATION: "You received new access request\n\n>*Resource Name*: `{{.resource_name}}`\n>*Access Level*: `{{.role}}`\n>*Requested for*: `{{.requestor}}`\n\nPlease visit https://console-beta.data.integration.golabs.io/dataaccess/manage-requests/{{.appeal_id}} to approve/reject the request."
AUTHENTICATED_USER_HEADER_KEY: X-Auth-Email
JOBS:
  # The Crontab time can be changed as per requirements
  FETCH_RESOURCES_INTERVAL: '0 */2 * * *' # default: "0 */2 * * *" which means "At minute 0 past every 2nd hour"
  REVOKE_EXPIRED_ACCESS_INTERVAL: '*/20 * * * *' # Default :"*/20 * * * *" meaning “At every 20th minute" 
  EXPIRING_ACCESS_NOTIFICATION_INTERVAL: '0 9 * * *' # Default:"0 9 * * *" meaning "At minute 0 past hour 9"
```

#### Note: 
The [Crontab schedule](https://crontab.guru) of Jobs is as per UTC timezone. With the above prompt, we have created a Job which fetches the resources every two hours. In case the access is revoked it will be checked every 20 minutes and lastly, if an access for a resources is getting expired, the user will be notified at 09:00 UTC everyday.

---

To initialize the database schema, Run Migrations with the following command:
```sh
$ guardian server migrate -c <path to the server.yml file>
```

To run the Guardian server use command:

```sh
$ guardian server start -c <path to the server.yml file>
```

---
### 2. Using environment variable

All the configs can be passed as environment variables using `<CONFIG_NAME>` convention. `<CONFIG_NAME>` is the key name of config using _ as the path delimiter to concatenate between keys. 

For example, to use environment variable, assuming the following configuration layout:

```
PORT : 8080
DB:
  HOST : localhost
  USER : test 
```
Here is the corresponding environment variable for the above

Configuration key | Environment variable |
------------------|----------------------|
PORT              | PORT                 |
DB.HOST           | DB_HOST              |
DB.USER           | DB_USER              |

Set the env variable using export
```
$ export PORT=8080
```



