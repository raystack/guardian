# GCS

Google Cloud Storage(in short GCS) is the object storage service offered by Google Cloud. GCS has distinct namespaces called **Buckets** that each one contains multiple **Objects** which are used for storing the data. 

It provides features such as object versioning or fine-grain permissions (per object or bucket). Using GCS one can retrieve and upload files using a REST API, and this can extend infinitely with each object scaling up to the terabyte size.

## Prerequisites

1. A service account with `roles/storage.admin` role at the project/organization level

## Access Management

Access can be given only at the bucket level on Guardian as those allowed to be managed through these Google Cloud Storage APIs:
- [Bucket Access Control](https://cloud.google.com/storage/docs/samples/storage-add-bucket-iam-member)


#### Provider Config
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

- user
- serviceAccount
- group
- domain

### GCS Credentials

| Fields | | |
| :--- | :--- | :--- |
| resource_name | string | GCP Project ID in resource name format. Example: `projects/my-project-id` |
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


