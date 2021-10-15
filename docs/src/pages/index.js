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
                    control centered around the type of data being accessed. RBAC is used
                    for authorizing people who can approve requests.
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
                title: 'Auditing for accountability',
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
                title: 'Access Control',
                content: (
                  <div>
                    Guardian provides a security framework through Data Platform IAM. It ensures
                    all users are authenticated with their credentials and are authorized to request,
                    approve, or perform any action on Guardian.
                  </div>
                ),
              },
              {
                title: 'Proven',
                content: (
                  <div>
                    Battle tested at large scale across multiple companies. Largest deployment manages
                    access from thousands of data sources.
                  </div>
                ),
              },
              {
                title: 'Analytics',
                content: (
                  <div>
                    Guardian provides continuous and real-time visibility by analyzing data usage by users,
                    groups, volume, tags, locations, and more. It allows users to generate reports about
                    instances of data access and related operations.
                  </div>
                ),
              },
            ]}
          />
        </Container>
        <Container className="textSection wrapper" background="dark">
          <h1>Framework</h1>
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
                title: 'Extraction',
                content: (
                  <div>
                    Extraction is the process of extracting data from a source and
                    transforming it into a format that can be consumed by the agent.
                    Extractors are the set of plugins that are source of our
                    metadata and include databases, dashboards, users, etc.
                  </div>
                ),
              },
              {
                title: 'Processing',
                content: (
                  <div>
                    Processing is the process of transforming the extracted data
                    into a format that can be consumed by the agent.
                    Processors are the set of plugins that perform the enrichment
                    or data processing for the metadata after extraction..
                  </div>
                ),
              },
              {
                title: 'Sink',
                content: (
                  <div>
                    Sink is the process of sending the processed data to a single or
                    multiple destinations as defined in recipes.
                    Sinks are the set of plugins that act as the destination of our metadata
                    after extraction and processing is done by agent.
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
                    title: 'Extractors',
                    content: (
                      <div>
                        Meteor supports source plugins to extract metadata from a variety of
                        datastores services, and message queues, including BigQuery,
                        InfluxDB, Kafka, Metabase, and many others.
                      </div>
                    ),
                  },
                  {
                    title: 'Processors',
                    content: (
                      <div>
                        Meteor has in-built processors inlcuding enrichment and others.
                        It is easy to add your own processors as well using custom plugins.
                      </div>
                    ),
                  },
                  {
                    title: 'Sinks',
                    content: (
                      <div>
                        Meteor supports sink plugins to send metadata to a variety of
                        third party APIs and catalog services, including Columbus, HTTP, BigQuery,
                        Kafka, and many others.
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
