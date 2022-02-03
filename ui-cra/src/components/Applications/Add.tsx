import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { ApplicationAdd, theme as weaveTheme } from '@weaveworks/weave-gitops';
import { useApplicationsCount } from './utils';
import styled from 'styled-components';

const ApplicationAddWrapper = styled(ApplicationAdd)`
  & .auth-message {
    margin-bottom: ${weaveTheme.spacing.xs};
  }
`;

const WGApplicationDetail: FC = () => {
  const applicationsCount = useApplicationsCount();

  return (
    <PageTemplate documentTitle="WeGO · Application Detail">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: applicationsCount,
          },
          { label: 'Add' },
        ]}
      />
      <ContentWrapper type="WG">
        <ApplicationAddWrapper />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationDetail;
