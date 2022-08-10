module.exports = {
  docsSidebar: [
    'introduction',
    'installation',
    'roadmap',
    {
      type: "category",
      label: "Concepts",
      items: [
        "concepts/overview",
      ],
    },
    {
      type: "category",
      label: "Guides",
      items: [
        "guides/introduction",
        "guides/configuration",
        "guides/create-policy",
        "guides/create-provider",
        "guides/update-resource",
        "guides/create-appeal",
        "guides/approve-reject-appeal",
        "guides/complex-use-case",
      ],
    },
    {
      type: "category",
      label: "Providers",
      items: [
        "providers/gcloud_iam",
        "providers/bigquery",
        "providers/gcs",
        "providers/tableau",
        "providers/metabase",
        "providers/grafana",
      ],
    },
    {
      type: "category",
      label: "Reference",
      items: [
        "reference/api",
        "reference/cli",
        "reference/appeal",
        "reference/policy",
        "reference/provider",
        "reference/resource",
        "reference/glossary",
      ],
    },
    {
      type: "category",
      label: "Contribute",
      items: [
        "contribute/architecture",
        "contribute/contribution",
        "contribute/development",
      ],
    },
  ],
};