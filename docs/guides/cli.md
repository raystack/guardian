# Guardian CLI

`Guardian` is a command line tool used to interact with the main guardian service.

## List of Commands

Guardian CLI supports many commands. To get a list of all the commands, follow these steps.

Enter the following code into the terminal:

```
$ guardian
```
The output is the following:
```
Available Commands:
  appeals     manage appeals
  config      manage guardian CLI configuration
  help        Help about any command
  migrate     Migrate database schema
  policies    manage policies
  providers   manage providers
  resources   manage resources
  serve       Run server
```

## Config command

Config command in Guardain's CLI is used to configure the command line tool. Following are a few examples of doing the same.

- ### What is inside?

Enter the following code into the terminal:

```
$ guardian config
```
The output is the following:
```
Available Commands:
  init        initialize CLI configuration
```

- ### init command

This command is used to initialize the `.guardian.yaml` file as demonstrated below.

Enter the following code into the terminal:

```
$ guardian config init
```
The output is the following:
```
config created: .guardian.yaml
```
Now, in the `.guardian.yaml` file we can set the configuartions as shown here.

```
host: localhost:3000
```

## Policies command

Policies command allows us to list, create or update policies.

- ### What is inside?

Enter the following code into the terminal:

```
$ guardian policies
```
The output is the following:
```
Available Commands:
  create      create policy
  list        list policies
  update      update policy
```
- ### create command

The create command is used to create a new policy. For this we have to define our policy file, which would be passed as a flag to the `create` command.

For instance, we can create a policy file `policy.yaml` as shown below.

```
id: policy_01
steps:
  - name: owner_approval
    approvers: $resource.details.owner
  - name: owner_approval2
    approvers: $resource.details.owner2
```

Now, we can create a policy using the `create` command as demonstrated here.

Enter the following code into the terminal:

```
$ guardian policies create --file policy.yaml
```
The output is the following:
```
policy created with id: policy_01  
```

- ### list command

To get a list of all the policies present in the Guardian' database, use the `list` command as explained here.

Enter the following code into the terminal:

```
$ guardian policies list
```
The output is the following:
```
  ID             VERSION  DESCRIPTION  STEPS                                            
  policy_10      1                     owner_approval                  
  policy_20      1                     owner_approval                  
  policy_01      1                     owner_approval,owner_approval2 
```

- ### update command

To update an existing policy present in the Guardian' database, use the `update` command as explained here.

For this first we update our `policy.yaml` file.

```
id: policy_01
steps:
  - name: owner_approval
    approvers: $resource.details.owner
  - name: admin_approval2
    approvers: $resource.details.admin
```

Enter the following code into the terminal:

```
$ guardian policies update --file policy.yaml
```
The output is the following:
```
policy updated
```

Note that on update of a policy it's version is also updated. We can verify this by listing all the policies.

```
  ID             VERSION  DESCRIPTION  STEPS                                            
  policy_10      1                     owner_approval                  
  policy_20      1                     owner_approval                  
  policy_01      2                     owner_approval,admin_approval 
```

## Providers command

Providers command allows us to list, create or update providers.

- ### What is inside?

Enter the following code into the terminal:

```
$ guardian providers
```
The output is the following:
```
Available Commands:
  create      register provider configuration
  list        list providers
  update      update provider configuration
```
- ### create command

The create command is used to create a new provider. For this we have to define our provider's config file, which would be passed as a flag to the `create` command.

For instance, we can create a config file `provider.yaml` for tableau provider as shown below.

```
type: tableau
urn: 691acb66-27ef-4b4f-9222-f07052e6ffg0
labels:
  entity: gojek
  landscape: id
credentials:
  host: https://prod-apnortheast-a.online.tableau.com
  username: user@test.com
  password: password@123
  content_url: guardiantestsite
appeal:
  allow_active_access_extension_in: 7d
resources:
  - type: metric
    policy:
      id: policy_20
      version: 1
    roles:
      - id: read
        name: Read
        permissions:
          - name: Read:Allow
      - id: write
        name: Write
        permissions:
          - name: Write:Allow
```

Now, we can create a provider using the `create` command as demonstrated here.

Enter the following code into the terminal:

```
$ guardian providers create --file provider.yaml
```
The output is the following:
```
provider created with id: 26
```

- ### list command

To get a list of all the providers present in the Guardian' database, use the `list` command as explained here.

Enter the following code into the terminal:

```
$ guardian providers list
```
The output is the following:
```
  ID  TYPE     URN                                   
  21  tableau  691acb66-27ef-4b4f-9222-f07052e6ffc2  
  22  tableau  691acb66-27ef-4b4f-9222-f07052e6ffc8  
  26  tableau  691acb66-27ef-4b4f-9222-f07052e6ffg0  
  24  tableau  691acb66-27ef-4b4f-9222-f07052e6ffd0  
```

- ### update command

To update an existing provider present in the Guardian' database, use the `update` command as explained here.

For this first we update our `provider.yaml` file.

After that, we can execute the update command as explained here.

Enter the following code into the terminal:

```
$ guardian providers update --file provider.yaml --id 26
```
The output is the following:
```
provider updated
``

## Resources command

Resources command allows us to list and set metadat of resoirces.

- ### What is inside?

Enter the following code into the terminal:

```
$ guardian resources
```
The output is the following:
```
Available Commands:
  list        list resources
  metadata    manage resource's metadata
```
- ### list command

It fetches the list of all the resources in the Guardian's database.

Enter the following code into the terminal:

```
$ guardian resources list
```
The output is the following:
```
ID    PROVIDER                              TYPE        URN                                   NAME                   
3552  tableau                               view        8a48df6d-bb5c-438f-a038-35149011e1b5  Flight Delays          
      691acb66-27ef-4b4f-9222-f07052e6ffc2                                                                           
4704  tableau                               metric      a408051f-c394-4a73-8f33-7bf7ba001d99  my-test-metric-ishan   
      691acb66-27ef-4b4f-9222-f07052e6ffc2                                                                           
3792  tableau                               workbook    7c940f8b-34c7-44af-9998-b95deef54edd  Regional               
      691acb66-27ef-4b4f-9222-f07052e6ffc8                                                                           
3802  tableau                               view        3bd3acd1-0681-458b-9566-0519ba844519  Overview               
      691acb66-27ef-4b4f-9222-f07052e6ffc8                                                                           
3807  tableau                               view        703c58f2-5b7f-46ba-bf96-9f4b473e4da8  Commission Model       
      691acb66-27ef-4b4f-9222-f07052e6ffc8                                                                           
5614  tableau                               view        7342fec1-4092-4bd4-abf4-8e531fe0f8ad  Stocks                 
      691acb66-27ef-4b4f-9222-f07052e6ffd0                                                                    
```

- ### metadata command

## Appeals command

Appeals command allows us to list, create, list and reject appeal.

- ### What is inside?

Enter the following code into the terminal:

```
$ guardian appeals
```
The output is the following:
```
Available Commands:
  approve     approve an approval step
  create      create appeal
  list        list appeals
  reject      reject an approval step
```
- ### create command

It helps us to create a new appeal.

Enter the following code into the terminal:

```
$ guardian appeals create --resource-id 5624 --role write --user test-user@email.com --options.duration "24h"
```
The output is the following:
```
appeal created with id: 13                                                                   
```

- ### list command

It helps us to get the list of all the appeals in the Guardian's database.

Enter the following code into the terminal:

```
$ guardian appeals list
```
The output is the following:
```
  ID  USER                  RESOURCE ID  ROLE   STATUS      
  11  ishanarya0@gmail.com  3807         write  pending     
  12  ishanarya0@gmail.com  5624         write  terminated  
  13  test-user@email.com   5624         write  pending                                                                
```

- ### approve command

It's used to approve an appeal.

- ### reject command

It's used to reject an appeal.




