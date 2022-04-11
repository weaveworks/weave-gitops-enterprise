import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
<<<<<<< HEAD
import { KustomizationDetail, useGetKustomization } from '@weaveworks/weave-gitops';

type Props = {
  name: string;
}
=======
<<<<<<< HEAD:ui-cra/src/components/Applications/Sources.tsx
import { SourcesTable, useListSources } from '@weaveworks/weave-gitops';

const WGApplicationsSources: FC = () => {
  const applicationsCount = useApplicationsCount();
  const { data: sources } = useListSources();

  return (
    <PageTemplate documentTitle="WeGO · Application Sources">
=======
import {
  KustomizationDetail,
  useGetKustomization,
} from '@weaveworks/weave-gitops';

type Props = {
  name: string;
};
>>>>>>> 9bb33d3d8bda881092f178c3b16c27df0b0a5b6b

const WGApplicationsKustomization: FC<Props> = ({ name }) => {
  const applicationsCount = useApplicationsCount();
  const { data } = useGetKustomization(name);
  const kustomization = data?.kustomization;

  return (
    <PageTemplate documentTitle="WeGO · Kustomization">
<<<<<<< HEAD
=======
>>>>>>> 9bb33d3d8bda881092f178c3b16c27df0b0a5b6b:ui-cra/src/components/Applications/Kustomization.tsx
>>>>>>> 9bb33d3d8bda881092f178c3b16c27df0b0a5b6b
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
<<<<<<< HEAD
        <KustomizationDetail kustomization={kustomization} name={name} />
=======
<<<<<<< HEAD:ui-cra/src/components/Applications/Sources.tsx
        <SourcesTable sources={sources} />
=======
        <KustomizationDetail kustomization={kustomization} name={name} />
>>>>>>> 9bb33d3d8bda881092f178c3b16c27df0b0a5b6b:ui-cra/src/components/Applications/Kustomization.tsx
>>>>>>> 9bb33d3d8bda881092f178c3b16c27df0b0a5b6b
      </ContentWrapper>
    </PageTemplate>
  );
};

<<<<<<< HEAD
export default WGApplicationsKustomization;
=======
<<<<<<< HEAD:ui-cra/src/components/Applications/Sources.tsx
export default WGApplicationsSources;
=======
export default WGApplicationsKustomization;
>>>>>>> 9bb33d3d8bda881092f178c3b16c27df0b0a5b6b:ui-cra/src/components/Applications/Kustomization.tsx
>>>>>>> 9bb33d3d8bda881092f178c3b16c27df0b0a5b6b
