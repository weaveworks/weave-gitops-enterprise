import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import { HelmChartDetail, useListSources } from '@weaveworks/weave-gitops';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};

const WGApplicationsHelmChart: FC<Props> = props => {
  const applicationsCount = useApplicationsCount();
  const { data: sources } = useListSources();

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
            count: sources?.length,
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
