# GCS

Google Cloud Storage(in short GCS) is the object storage service offered by Google Cloud. GCS has distinct namespaces called **Buckets** that each one contains multiple **Objects** which are used for storing the data.

It provides features such as object versioning or fine-grain permissions (per object or bucket). Using GCS one can retrieve and upload files using a REST API, and this can extend infinitely with each object scaling up to the terabyte size.

### GCS Resources

- **Organization**: The Organization resource represents an organization (for example, a
  company) and is the root node in the Google Cloud resource hierarchy. Your
  company, called Example Inc., creates a Google Cloud organization called
  exampleinc.org.
- **Project**: A project organizes all your Google Cloud resources. All data in Cloud Storage
  belongs inside a project. A project consists of a set of users; a set of APIs; and billing,
  authentication, and monitoring settings for those APIs. Example Inc. is building several
  applications, and each one is associated with a project.
- **Bucket**: Each project can contain multiple buckets, which are containers to store your
  objects.You can use buckets to organize your data and control access to your data, but
  unlike directories and folders, you cannot nest buckets. For example, you might create a
  photos bucket for all the image files your app generates and a separate video bucket.
- **Object**: Objects are the individual pieces of data that you store in Cloud Storage.There is
  no limit on the number of objects that you can create in a bucket. An individual file, such as an image called raystack.png.

### GCS Users

GCS allows Google Account, Service account, Google group, Google Workspace account, Cloud
Identity domain, All authenticated users, All users allowed to access the buckets and objects
inside a specific bucket.

Currently, Guardian supports **`user`**, **`service account`**, **`group`** and **`domain`** as
allowed account types.

### Prerequisites

If a user/administrator wants to control access to a bucket or an object, the user must have
sufficient permissions for the same. With these permissions, the resource owner can grant and
revoke other users/service accounts with selective access to these resources.

For registering Google Cloud Storage as a provider on Guardian, user must have a service
account with **`roles/storage.admin`** role at the project/organization level

### Authentication

Guardian requires a **service account key** and the **resource name** of an administrator user in
Google Cloud Storage. The Service Account key should be base64 encoded value.

```yaml
credentials:
Service_account_key: <base64 encoded Service Account Key>
Resource_name: projects/gcs-project-i
```

## Access Management

Access can be given only at the bucket level on Guardian as those allowed to be managed through these Google Cloud Storage APIs:

- [Bucket Access Control](https://cloud.google.com/storage/docs/samples/storage-add-bucket-iam-member)

## Provider Config

#### YAML Representation

```yaml
type: gcs
urn: sample-URN
credentials:
  service_account_key: <base64 encoded service account key json>
  resource_name: projects/<gcs-project-id>
  - type: bucket
    policy:
      id: my-first-policy
      version: 1
    roles:
      - id: READER
        name: Reader
        description: 'Grants permission to list a bucket contents and read bucket metadata, excluding IAM policies'
        permissions:
          - roles/storage.legacyBucketReader
      - id: WRITER
        name: Writer
        description: 'Grants permission to create, replace, and delete objects; list objects in a bucket'
        permissions:
          - roles/storage.legacyBucketWriter
      - id: OWNER
        name: Owner
        description: 'Grants permission to update objects; list and update tag bindings; read object metadata when listing'
        permissions:
          - roles/storage.legacyBucketOwner
      - id: ADMIN
        name: Admin
        description: 'Grants full control of buckets and objects'
        permissions:
          - roles/storage.admin
      - id: OBJECTADMIN
        name: ObjectAdmin
        description: 'Grants full control over objects, including listing, creating, viewing, and deleting objects'
        permissions:
          - roles/storage.objectAdmin
```

### GCS Account Types

- `user`
- `serviceAccount`
- `group`
- `domain`

### GCS Credentials

| Fields              |        |                                                                                                                                                           |
| :------------------ | :----- | :-------------------------------------------------------------------------------------------------------------------------------------------------------- |
| resource_name       | string | GCP Project ID in resource name format. Example: `projects/my-project-id`                                                                                 |
| service_account_key | string | Service account key JSON that has [prerequisites permissions](#prerequisites).<br/> On provider creation, the value should be an base64 encoded JSON key. |

### GCS Resource Types

- Bucket

### GCS Resource Permission

A Google Cloud predefined role name. [`Read More`](https://cloud.google.com/storage/docs/access-control/iam-roles)

For **`Bucket`** resource type, the list of allowed permissions are:

- `roles/storage.admin`
- `roles/storage.legacyBucketOwner`
- `roles/storage.legacyBucketReader`
- `roles/storage.legacyBucketWriter`
- `roles/storage.legacyObjectOwner`
- `roles/storage.legacyObjectReader`
- `roles/storage.objectAdmin`
- `roles/storage.objectCreator`
- `roles/storage.objectViewer`
