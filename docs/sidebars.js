module.exports = {
  docsSidebar: [
    "introduction",
    "installation",
    "roadmap",
    {
      type: "category",
      label: "Tour",
      link: {
        type: "doc",
        id: "tour/introduction",
      },
      items: [
        "tour/introduction",
        "tour/configuration",
        "tour/create-policy",
        "tour/create-provider",
        "tour/update-resource",
        "tour/create-appeal",
        "tour/approve-reject-appeal",
        "tour/complex-use-case",
      ],
    },
    {
      type: "category",
      label: "Concepts",
      link: {
        type: "doc",
        id: "concepts/overview",
      },
      items: ["concepts/overview", "concepts/architecture"],
    },
    {
      type: "category",
      label: "Guides",
      link: {
        type: "doc",
        id: "guides/deployment",
      },
      items: ["guides/deployment"],
    },
    {
      type: "category",
      label: "Providers",
      link: {
        type: "generated-index",
        title: "Overview",
        slug: "providers/overview",
      },
      items: [
        "providers/bigquery",
        "providers/gcloud_iam",
        "providers/gcs",
        "providers/grafana",
        "providers/metabase",
        "providers/noop",
        "providers/tableau",
        "providers/frontier",
      ],
    },
    {
      type: "category",
      label: "Reference",
      link: {
        type: "generated-index",
        title: "Overview",
        slug: "reference/overview",
      },
      items: [
        "reference/api",
        "reference/cli",
        "reference/appeal",
        "reference/policy",
        "reference/provider",
        "reference/resource",
        "reference/jobs",
        "reference/glossary",
        "reference/configuration",
      ],
    },
    {
      type: "category",
      label: "APIs",
      link: {
        type: "doc",
        id: "apis/guardian-apis",
      },
      items: require("./docs/apis/sidebar.js"),
    },
    {
      type: "category",
      label: "Extend",
      link: {
        type: "doc",
        id: "contribute/provider",
      },
      items: ["contribute/provider"],
    },
    {
      type: "category",
      label: "Contribute",
      link: {
        type: "doc",
        id: "contribute/contribution",
      },
      items: ["contribute/architecture", "contribute/contribution"],
    },
  ],
};
