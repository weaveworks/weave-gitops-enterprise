import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount, useSourcesCount } from './utils';
import { HelmChartDetail } from '@weaveworks/weave-gitops';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};

const WGApplicationsHelmChart: FC<Props> = props => {
  const applicationsCount = useApplicationsCount();
  const sourcesCount = useSourcesCount();

  return (
    <PageTemplate documentTitle="WeGO · Helm Chart">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: applicationsCount,
          },
          {
            label: 'Sources',
            url: '/sources',
            count: sourcesCount,
          },
          {
            label: `${props.name}`,
          },
        ]}
      />
      <ContentWrapper>
        <HelmChartDetail {...props} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsHelmChart;
