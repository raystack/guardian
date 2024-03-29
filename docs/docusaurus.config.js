const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
(module.exports = {
  title: 'Guardian',
  tagline: 'Universal data access control',
  url: 'https://guardian.raystack.org/',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  organizationName: 'Raystack',
  projectName: 'guardian',

  presets: [
    [
      '@docusaurus/preset-classic',
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/raystack/guardian/edit/master/docs/',
          sidebarCollapsed: true,
          breadcrumbs: false,
          docLayoutComponent: "@theme/DocPage",
          docItemComponent: "@theme/ApiItem" 
        },
        blog: false,
        theme: {
          customCss: [
            require.resolve('./src/css/theme.css'),
            require.resolve('./src/css/custom.css')
          ],
        },
        gtag: {
          trackingID: 'G-EPXDLH6V72',
        },
      }),
    ],
  ],
  plugins: [
    [
      'docusaurus-plugin-openapi-docs',
      {
        id: "apiDocs",
        docsPluginId: "classic",
        config: {
          auth: {
            specPath: "../api/proto/raystack/guardian/v1beta1/guardian.swagger.json",
            outputDir: "docs/apis",
            sidebarOptions: {
              groupPathsBy: "tag",
            },
            hideSendButton: false,
          }
        }
      },
    ],
  ],
  themes: ["docusaurus-theme-openapi-docs"],
  themeConfig:
    ({
      image: 'img/banner.png',
      colorMode: {
        defaultMode: 'light',
        respectPrefersColorScheme: true,
      },
      docs: {
        sidebar: {
          autoCollapseCategories: true,
          hideable: true,
        },
      },
      navbar: {
        title: 'Guardian',
        logo: { src: 'img/logo.svg', },
        hideOnScroll: true,
        items: [
          {
            type: 'doc',
            docId: 'introduction',
            position: 'right',
            label: 'Docs',
          },
          { to: 'docs/support', label: 'Support', position: 'right' },
          {
            href: 'https://bit.ly/2RzPbtn',
            position: 'right',
            className: 'header-slack-link',
          },
          {
            href: 'https://github.com/raystack/guardian',
            className: 'navbar-item-github',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'light',
        links: [],
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
      announcementBar: {
        id: 'star-repo',
        content: '⭐️ If you like Guardian, give it a star on <a target="_blank" rel="noopener noreferrer" href="https://github.com/raystack/guardian">GitHub</a>! ⭐',
        backgroundColor: '#222',
        textColor: '#eee',
        isCloseable: true,
      },
    }),
});
