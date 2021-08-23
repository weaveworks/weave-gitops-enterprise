import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import useClusters from '../../contexts/Clusters';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';

const TemplatesDashboard: FC = ({ children }) => {
  const clustersCount = useClusters().count;

  return (
    <PageTemplate documentTitle="WeGO · Templates">
      <SectionHeader
        path={[{ label: 'Clusters', url: '/clusters', count: clustersCount }]}
      />
      <ContentWrapper>{children}</ContentWrapper>
    </PageTemplate>
  );
};

export default TemplatesDashboard;
