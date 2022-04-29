import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import { HelmChartDetail } from '@weaveworks/weave-gitops';

type Props = {
  name: string;
  namespace: string;
};

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
        <HelmChartDetail name={name} namespace={namespace} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsHelmChart;
