import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import { HelmRepositoryDetail, useListSources } from '@weaveworks/weave-gitops';

type Props = {
  name: string;
  namespace: string;
};

const WGApplicationsHelmRepository: FC<Props> = props => {
  const applicationsCount = useApplicationsCount();
  const { data: sources } = useListSources();

  return (
    <PageTemplate documentTitle="WeGO · Helm Repository">
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
        <Title>{props.name}</Title>
        <HelmRepositoryDetail {...props} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsHelmRepository;
