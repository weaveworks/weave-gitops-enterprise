import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import { Timestamp, Interval, SourceRefSourceKind, SourceDetail } from '@weaveworks/weave-gitops';

type Props = {
  name: string;
  namespace: string;
}

const WGApplicationsBucket: FC<Props> = ({ name, namespace }) => {
  const applicationsCount = useApplicationsCount();

  return (
    <PageTemplate documentTitle="WeGO · Bucket">
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
          type={SourceRefSourceKind.Bucket}
          // Guard against an undefined bucket with a default empty object
          info={(b: any = {}) => [
            ["Endpoint", b.endpoint],
            ["Bucket Name", b.name],
            ["Last Updated", <Timestamp time={b.lastUpdatedAt} />],
            ["Interval", <Interval interval={b.interval} />],
            ["Cluster", "Default"],
            ["Namespace", b.namespace],
          ]}
        />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsBucket;
