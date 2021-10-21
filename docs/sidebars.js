module.exports = {
  docsSidebar: [
    'introduction',
    {
      type: "category",
      label: "Guides",
      items: [
        "guides/overview",
        "guides/cli",
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
        "providers/tableau",
        "providers/metabase",
        "providers/grafana",
      ],
    },
    {
      type: "category",
      label: "Concepts",
      items: [
        "concepts/architecture",
      ],
    },
    {
      type: "category",
      label: "Reference",
      items: [
        "reference/glossary",
        "reference/policy-config",
        "reference/provider-config",
      ],
    },
    {
      type: "category",
      label: "Contribute",
      items: [
        "contribute/contribution",
        "contribute/development",
      ],
    },
    'roadmap'

  ],
};