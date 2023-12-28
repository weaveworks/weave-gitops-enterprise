import { SourcesTable, useListSources } from '@weaveworks/weave-gitops';
import { FC, useEffect } from 'react';
import styled from 'styled-components';
import { EnabledComponent } from '../../api/query/query.pb';
import useNotifications from '../../contexts/Notifications';
import { useIsEnabledForComponent } from '../../hooks/query';
import { formatError } from '../../utils/formatters';
import Explorer from '../Explorer/Explorer';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';

const WGApplicationsSources: FC = ({ className }: any) => {
  const isExplorerEnabled = useIsEnabledForComponent(EnabledComponent.sources);

  const {
    data: sources,
    isLoading,
    error,
  } = useListSources('', '', {
    enabled: !isExplorerEnabled,
    retry: false,
    refetchInterval: 5000,
  });

  const { setNotifications } = useNotifications();

  useEffect(() => {
    if (error) {
      setNotifications(formatError(error));
    }
  }, [error, setNotifications]);

  return (
    <Page
      loading={!isExplorerEnabled && isLoading}
      path={[
        {
          label: 'Sources',
        },
      ]}
    >
      <NotificationsWrapper errors={sources?.errors}>
        <div className={className}>
          {isExplorerEnabled ? (
            <Explorer enableBatchSync category="source" />
          ) : (
            <SourcesTable sources={sources?.result} />
          )}
        </div>
      </NotificationsWrapper>
    </Page>
  );
};

export default styled(WGApplicationsSources)``;
