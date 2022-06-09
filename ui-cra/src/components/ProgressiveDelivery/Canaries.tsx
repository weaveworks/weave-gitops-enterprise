import { Alert } from '@material-ui/lab';
import { FilterableTable, LoadingPage } from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import { useApplicationsCount } from '../Applications/utils';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { useListCanaries } from './hooks';

type Props = {
  className?: string;
};

function Canaries({ className }: Props) {
  const applicationsCount = useApplicationsCount();
  const { data: canaries, error, isLoading } = useListCanaries();

  if (isLoading) {
    return <LoadingPage />;
  }

  return (
    <div className={className}>
      <PageTemplate documentTitle="WeGO · Canaries">
        <SectionHeader
          path={[
            {
              label: 'Applications',
              url: '/applications',
              count: applicationsCount,
            },
            {
              label: 'Canaries',
              url: '/delivery',
              count: 0,
            },
          ]}
        />
        <ContentWrapper>
          <Title>Canaries</Title>
          {error && <Alert severity="error">{error.message}</Alert>}
          <FilterableTable
            filters={{}}
            fields={[{ value: 'name', label: 'Name' }]}
            rows={canaries?.canaries || []}
          />
        </ContentWrapper>
      </PageTemplate>
    </div>
  );
}

export default styled(Canaries).attrs({ className: Canaries.name })``;
