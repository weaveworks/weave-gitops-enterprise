import {
  SourcesTable,
  useListSources,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { Routes } from '../../utils/nav';
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
          url: Routes.Applications,
        },
        {
          label: 'Sources',
          url: V2Routes.Sources,
        },
      ]}
    >
      <ContentWrapper
        errors={sources?.errors}
        loading={isLoading}
        notification={{
          message: { text: error?.message },
          severity: 'error',
        }}
      >
        {sources && <SourcesTable sources={sources?.result} />}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsSources;
