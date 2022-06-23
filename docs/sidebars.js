module.exports = {
  docsSidebar: [
    'introduction',
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
        "guides/cli",
        "guides/api",
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
        "reference/appeal",
        "reference/policy",
        "reference/provider",
        "reference/resource",
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