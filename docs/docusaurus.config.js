const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

// With JSDoc @type annotations, IDEs can provide config autocompletion
/** @type {import('@docusaurus/types').DocusaurusConfig} */
(module.exports = {
  title: 'Guardian',
  tagline: 'Universal data access control',
  url: 'https://goto.github.io/',
  baseUrl: '/guardian/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  organizationName: 'GOTO',
  projectName: 'guardian',

  presets: [
    [
      '@docusaurus/preset-classic',
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/goto/guardian/edit/master/docs/',
          sidebarCollapsed: true,
          breadcrumbs: false,
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

  themeConfig:
    ({
      colorMode: {
        defaultMode: 'light',
        respectPrefersColorScheme: true,
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
            href: 'https://github.com/goto/guardian',
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
        content: '⭐️ If you like Guardian, give it a star on <a target="_blank" rel="noopener noreferrer" href="https://github.com/goto/guardian">GitHub</a>! ⭐',
        backgroundColor: '#222',
        textColor: '#eee',
        isCloseable: true,
      },
    }),
});
