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
};

const WGApplicationsKustomization: FC<Props> = ({ name }) => {
  const applicationsCount = useApplicationsCount();
  const { data } = useGetKustomization(name);
  const kustomization = data?.kustomization;

  return (
    <PageTemplate documentTitle="WeGO · Kustomization">
      <SectionHeader
        path={[
          {
            label: 'Sources',
            url: '/sections',
            count: applicationsCount,
          },
        ]}
      />
      <ContentWrapper type="WG">
        <KustomizationDetail kustomization={kustomization} name={name} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsKustomization;
