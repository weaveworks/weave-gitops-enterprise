import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import {
  KustomizationDetail,
  useGetKustomization,
} from '@weaveworks/weave-gitops';

type Props = {
  name: string;
  namespace?: string;
  clusterName: string;
};

const WGApplicationsKustomization: FC<Props> = ({
  name,
  namespace,
  clusterName,
}) => {
  const applicationsCount = useApplicationsCount();
  const { data } = useGetKustomization(name, namespace, clusterName);
  const kustomization = data?.kustomization;

  return (
    <PageTemplate documentTitle="WeGO · Kustomization">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: applicationsCount,
          },
          {
            label: `${name}`,
          },
        ]}
      />
      <ContentWrapper>
        <KustomizationDetail kustomization={kustomization} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsKustomization;
