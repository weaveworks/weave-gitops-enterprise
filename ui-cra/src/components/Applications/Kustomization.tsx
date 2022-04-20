import React, { FC, useContext } from 'react';
import styled from 'styled-components';

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

const KustomizationDetailWrapper = styled(KustomizationDetail)`
  div[class^='ReconciliationGraph'],
  div[class*='ReconciliationGraph'] {
    svg {
      min-height: 600px;
    }
    .MuiSlider-root.MuiSlider-vertical {
      height: 200px;
    }
  }
`;

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
      <ContentWrapper>
        <KustomizationDetailWrapper kustomization={kustomization} name={name} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsKustomization;
