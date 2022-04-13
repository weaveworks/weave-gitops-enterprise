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

const WGApplicationsHelmChart: FC<Props> = ({ name, namespace }) => {
  const applicationsCount = useApplicationsCount();

  return (
    <PageTemplate documentTitle="WeGO · Helm Chart">
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
          type={SourceRefSourceKind.HelmChart}
          info={(ch: any) => [
            ["Chart", ch?.chart],
            ["Ref", ch?.sourceRef?.name],
            ["Last Updated", <Timestamp time={ch?.lastUpdatedAt} />],
            ["Interval", <Interval interval={ch?.interval} />],
            ["Namespace", ch?.namespace],
          ]}
        />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsHelmChart;
