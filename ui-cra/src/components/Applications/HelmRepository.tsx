import React, { FC } from 'react';
import { NameLink } from '../Shared';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import { Interval, Timestamp, SourceRefSourceKind, SourceDetail } from '@weaveworks/weave-gitops';

type Props = {
  name: string;
  namespace: string;
}

const WGApplicationsHelmRepository: FC<Props> = ({ name, namespace }) => {
  const applicationsCount = useApplicationsCount();

  return (
    <PageTemplate documentTitle="WeGO · Helm Repository">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: applicationsCount,
          },
        ]}
      />
      <ContentWrapper type="WG">
        <SourceDetail
          name={name}
          namespace={namespace}
          type={SourceRefSourceKind.HelmRepository}
          info={(hr: any = {}) => [
            [
              "URL",
              <NameLink href={hr.url}>
                {hr.url}
              </NameLink>,
            ],
            ["Last Updated", <Timestamp time={hr.lastUpdatedAt} />],
            ["Interval", <Interval interval={hr.interval} />],
            ["Cluster", "Default"],
            ["Namespace", hr.namespace],
          ]}
        />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsHelmRepository;
