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
        },
        blog: {
          showReadingTime: true,
          editUrl:
            'https://github.com/odpf/guardian/edit/master/docs/blog/',
        },
        theme: {
          customCss: [
            require.resolve('./src/css/theme.css'),
            require.resolve('./src/css/custom.css')
          ],
        },
      }),
    ],
  ],

  themeConfig:
    ({
      colorMode: {
        defaultMode: 'light',
        respectPrefersColorScheme: true,
        switchConfig: {
          darkIcon: '☾',
          lightIcon: '☀️',
        },
      },
      navbar: {
        title: 'Guardian',
        logo: { src: 'img/logo.svg', },
        items: [
          {
            type: 'doc',
            docId: 'introduction',
            position: 'left',
            label: 'Docs',
          },
          { to: '/blog', label: 'Blog', position: 'left' },
          { to: '/help', label: 'Help', position: 'left' },
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
      gtag: {
        trackingID: 'G-ZTPBZN6VK7',
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
