import { SourcesTable, useListSources } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';

const WGApplicationsSources: FC = () => {
  const { data: sources, isLoading, error } = useListSources();
  return (
    <PageTemplate
      documentTitle="Application Sources"
      path={[
        {
          label: 'Applications',
          url: '/applications',
        },
        {
          label: 'Sources',
          url: '/sources',
          count: sources?.result?.length,
        },
      ]}
    >
      <ContentWrapper
        errors={sources?.errors}
        loading={isLoading}
        errorMessage={error?.message}
      >
        {sources && <SourcesTable sources={sources?.result} />}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsSources;
