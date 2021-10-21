import React from 'react';
import Layout from '@theme/Layout';
import clsx from 'clsx';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Container from '../core/Container';
import GridBlock from '../core/GridBlock';
import useBaseUrl from '@docusaurus/useBaseUrl';

const Hero = () => {
  const { siteConfig } = useDocusaurusContext();
  return (
    <div className="homeHero">
      <div className="logo"><img src={useBaseUrl('img/pattern.svg')} /></div>
      <div className="container banner">
        <div className="row">
          <div className={clsx('col col--5')}>
            <div className="homeTitle">{siteConfig.tagline}</div>
            <small className="homeSubTitle">Guardian is a tool for extensible and universal data access with automated access workflows and security controls across data stores, analytical systems, and cloud products.</small>
            <a className="button" href="docs/introduction">Documentation</a>
          </div>
          <div className={clsx('col col--1')}></div>
          <div className={clsx('col col--6')}>
            <div className="text--right"><img src={useBaseUrl('img/banner.svg')} /></div>
          </div>
        </div>
      </div>
    </div >
  );
};

export default function Home() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <Layout
      title={siteConfig.tagline}
      description="Meteor is an easy-to-use, plugin-driven metadata collection framework to extract data from different sources and sink to any data catalog or store.">
      <Hero />
      <main>
        <Container className="textSection wrapper" background="light">
          <h1>Built for security</h1>
          <p>
            Guardian is the data access and control solution,
            enabling data teams to accelerate data delivery,
            reduce risk, and safely unlock more data.
          </p>
          <GridBlock
            layout="threeColumn"
            contents={[
              {
                title: 'Hybrid access control',
                content: (
                  <div>
                    Guardian uses a hybrid approach of role-based and data-centric access
                    control centered around the type of resource being accessed. It also provides
                    time-limited access to ensuring security and compliance.
                  </div>
                ),
              },
              {
                title: 'Compliant workflows',
                content: (
                  <div>
                    Guardian decouples access controls from data stores to allow for better
                    integration without disrupting your existing data workflow. It also
                    provides custom workflows to make compliance easy for any framework,
                    from GDPR to HITRUST.
                  </div>
                ),
              },
              {
                title: 'Auditing',
                content: (
                  <div>
                    Automatically produce and maintain complete, interpretable records of
                    data access activities to track every operation. It allows auditors to track
                    requests and access to data, policy changes, how information is being used,
                    and more.
                  </div>
                ),
              },
              {
                title: 'Management',
                content: (
                  <div>
                    Guardian comes with CLI and APIs which allows you to interact with access
                    workflows effectively. You can manage resources, providers, appeals, and policies
                    and more.
                  </div>
                ),
              },
              {
                title: 'Proven',
                content: (
                  <div>
                    Guardian is battle tested at large scale across multiple companies. Largest
                    deployment manages access for thousands of resources across different providers.
                  </div>
                ),
              },
              {
                title: 'Analytics',
                content: (
                  <div>
                    Guardian provides continuous and real-time visibility by analyzing access usage by users,
                    groups, and more. It generates reports about instances of data
                    access and related operations.
                  </div>
                ),
              },
            ]}
          />
        </Container>
        <Container className="textSection wrapper" background="dark">
          <h1>Key Features</h1>
          <p>
            Meteor agent uses recipes as a set of instructions which are configured by user.
            Recipes contains configurations about the source from which the metadata will be
            fetched, information about metadata processors and the destination to where
            the metadata will be sent.
          </p>
          <GridBlock
            layout="threeColumn"
            contents={[

              {
                title: 'Appeal-based access',
                content: (
                  <div>
                    Users are expected to create an appeal for accessing data from registered
                    providers. The appeal will get reviewed by the configured approvers before
                    it gives the access to the user.
                  </div>
                ),
              },
              {
                title: 'Configurable approval flow',
                content: (
                  <div>
                    Approval flow configures what are needed for an appeal to get approved
                    and who are eligible to approve/reject. It can be configured and linked
                    to a provider so that every appeal created to their resources will follow
                    the procedure in order to get approved.
                  </div>
                ),
              },
              {
                title: 'External Identity Manager',
                content: (
                  <div>
                    Guardian gives the flexibility to use any third-party identity manager for user properties.
                  </div>
                ),
              },
            ]}
          />
        </Container>


        <Container className="textSection wrapper" background="light">
          <h1>Ecosystem</h1>
          <p>
            Meteor’s plugin system allows new plugins to be easily added.
            With 50+ plugins and many more coming soon to extract and sink metadata,
            it is easy to start collecting metadata from various sources and
            sink to any data catalog or store.
          </p>
          <div className="row">
            <div className="col col--4">

              <GridBlock
                contents={[
                  {
                    title: 'Providers',
                    content: (
                      <div>
                        Support various providers like Big Query, Metabase, Tableau,
                        and multiple instances for each provider type.
                      </div>
                    ),
                  },
                  {
                    title: 'Resources',
                    content: (
                      <div>
                        Resources from a provider are managed in Guardian's database.
                        There is also an API to update resource's metadata to add additional information.
                      </div>
                    ),
                  },
                  {
                    title: 'Appeals',
                    content: (
                      <div>
                        Appeal is created by a user with specifying which resource they want
                        to access along with some other appeal options.
                      </div>
                    ),
                  },
                ]}
              />
            </div>
            <div className="col col--8">
              <img src={useBaseUrl('assets/overview.svg')} />
            </div>
          </div>
        </Container>

        {/* <Container className="textSection wrapper" background="light">
          <h1>Trusted by</h1>
          <p>
            Meteor was originally created for the Gojek data processing platform,
            and it has been used, adapted and improved by other teams internally and externally.
          </p>
          <GridBlock className="logos"
            layout="fourColumn"
            contents={[
              {
                content: (
                  <img src={useBaseUrl('users/gojek.png')} />
                ),
              },
              {
                content: (
                  <img src={useBaseUrl('users/midtrans.png')} />
                ),
              },
              {
                content: (
                  <img src={useBaseUrl('users/mapan.png')} />
                ),
              },
              {
                content: (
                  <img src={useBaseUrl('users/moka.png')} />
                ),
              },
            ]}>
          </GridBlock>
        </Container> */}
      </main>
    </Layout >
  );
}
