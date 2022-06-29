const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

// With JSDoc @type annotations, IDEs can provide config autocompletion
/** @type {import('@docusaurus/types').DocusaurusConfig} */
(module.exports = {
  title: 'Guardian',
  tagline: 'Universal data access control',
  url: 'https://odpf.github.io/',
  baseUrl: '/guardian/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  organizationName: 'ODPF',
  projectName: 'guardian',

  presets: [
    [
      '@docusaurus/preset-classic',
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/odpf/guardian/edit/master/docs/',
          sidebarCollapsed: false,
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
            docId: 'overview/introduction',
            position: 'right',
            label: 'Documentation',
          },
          { to: '/help', label: 'Support', position: 'right' },
          {
            href: 'https://bit.ly/2RzPbtn',
            position: 'right',
            className: 'header-slack-link',
          },
          {
            href: 'https://github.com/odpf/guardian',
            className: 'navbar-item-github',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'light',
        links: [
          {
            title: 'Products',
            items: [
              { label: 'Optimus', href: 'https://github.com/odpf/optimus' },
              { label: 'Firehose', href: 'https://github.com/odpf/firehose' },
              { label: 'Raccoon', href: 'https://github.com/odpf/raccoon' },
              { label: 'Dagger', href: 'https://odpf.github.io/dagger/' },
            ],
          },
          {
            title: 'Resources',
            items: [
              { label: 'Docs', to: '/docs/introduction' },
              { label: 'Blog', to: '/blog', },
              { label: 'Help', to: '/help', },
            ],
          },
          {
            title: 'Community',
            items: [
              { label: 'Slack', href: 'https://bit.ly/2RzPbtn' },
              { label: 'GitHub', href: 'https://github.com/odpf/guardian' }
            ],
          },
        ],
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
      announcementBar: {
        id: 'star-repo',
        content: '⭐️ If you like Guardian, give it a star on <a target="_blank" rel="noopener noreferrer" href="https://github.com/odpf/guardian">GitHub</a>! ⭐',
        backgroundColor: '#222',
        textColor: '#eee',
        isCloseable: true,
      },
    }),
});
