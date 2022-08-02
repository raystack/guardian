# GCS

Google Cloud Storage(in short GCS) is the object storage service offered by Google Cloud. GCS has distinct namespaces called **Buckets** that each one contains multiple **Objects** which are used for storing the data. 

It provides features such as object versioning or fine-grain permissions (per object or bucket). Using GCS one can retrieve and upload files using a REST API, and this can extend infinitely with each object scaling up to the terabyte size.

## Prerequisites

1. A service account with `` role at the project/organization level

## Access Management

Access can be given at the bucket level or object level as those allowed to be managed through these Google Cloud Storage APIs:
- [Bucket Access Control](https://cloud.google.com/storage/docs/samples/storage-add-bucket-iam-member)
- [Object Access Control](https://cloud.google.com/storage/docs/samples/storage-add-file-owner)

** NOTE ** : Object level access can only be provide for objects belonging to a **Fine-granied** bucket. For objects belonging to the **Uniform-level**, permissions are only given at the bucket level.

## Config
TODO

## GCS Account Types

- user
- serviceAccount
- group
- domain

## GCS Credentials

| Fields | | |
| :--- | :--- | :--- |
| resource_name | string | GCP Project ID in resource name format. Example: `projects/my-project-id` |
| service_account_key | string | Service account key JSON that has [prerequisites permissions](#prerequisites).<br/> On provider creation, the value should be an base64 encoded JSON key. |

## GCS Resource Types

- Bucket
- Object

## GCS Resource Permission

A Google Cloud predefined role name. 

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

For **`Object`** resource type, the list of allowed permissions are:

TODO

