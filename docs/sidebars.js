module.exports = {
  docsSidebar: [
    {
      type: "category",
      label: "Overview",
      items: [
        "overview/introduction",
        "overview/roadmap",
      ],
    },
    {
      type: "category",
      label: "Getting Started",
      items: [
        "getting_started/installation",
        "getting_started/configuration",
      ],
    },
    {
      type: "category",
      label: "Guides",
      items: [
        "guides/introduction",
        "guides/overview",
        "guides/create-policy",
        "guides/create-provider",
        "guides/update-resource",
        "guides/create-appeal",
        "guides/approve-reject-appeal",
        "guides/complex-use-case",
        "guides/managing-policies",
        "guides/managing-providers",
        "guides/managing-resources",
        "guides/managing-appeals",
      ],
    },
    {
      type: "category",
      label: "Providers",
      items: [
        "providers/bigquery",
        "providers/gcloud_iam",
        "providers/tableau",
        "providers/metabase",
        "providers/grafana",
      ],
    },
    // {
    //   type: "category",
    //   label: "Concepts",
    //   items: [
    //   ],
    // },
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