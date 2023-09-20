import { Template } from '../../cluster-services/cluster_services.pb';
import useNotifications, {
  NotificationData,
} from '../../contexts/Notifications';
import useTemplates from '../../hooks/templates';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import {
  Button,
  DataTable,
  Icon,
  IconType,
  Link,
  filterConfig,
  useFeatureFlags,
} from '@weaveworks/weave-gitops';
import { FC, useCallback, useEffect } from 'react';
import { useHistory } from 'react-router-dom';
import Explorer from '../Explorer/Explorer';

const Error = styled.span`
  color: ${props => props.theme.colors.alertOriginal};
`;
const CustomEmptyMessage = styled.span`
  color: ${props => props.theme.colors.neutral30};
`;
const DocsLink = styled(Link)`
  color: ${props => props.theme.colors.primary};
  padding-right: ${({ theme }) => theme.spacing.xxs};
  padding-left: ${({ theme }) => theme.spacing.xxs};
`;

const TemplatesDashboard: FC<{
  location: { state: { notification: NotificationData[] } };
}> = ({ location }) => {
  const { isFlagEnabled } = useFeatureFlags();
  const useQueryServiceBackend = isFlagEnabled(
    'WEAVE_GITOPS_FEATURE_QUERY_SERVICE_BACKEND',
  );
  const { templates, isLoading } = useTemplates({
    enabled: !useQueryServiceBackend,
  });
  const { setNotifications } = useNotifications();
  const history = useHistory();

  const initialFilterState = {
    ...filterConfig(templates, 'provider'),
    ...filterConfig(templates, 'namespace'),
    ...filterConfig(templates, 'templateType'),
  };

  const handleAddCluster = useCallback(
    (event, t) =>
      history.push(`/templates/create?name=${t.name}&namespace=${t.namespace}`),
    [history],
  );

  useEffect(
    () =>
      setNotifications([
        {
          message: {
            text: location?.state?.notification?.[0]?.message.text,
          },
          severity: location?.state?.notification?.[0]?.severity,
        } as NotificationData,
      ]),
    [location?.state?.notification, setNotifications],
  );

  return (
    <Page
      loading={isLoading}
      path={[
        {
          label: 'Templates',
        },
      ]}
    >
      <NotificationsWrapper>
        {useQueryServiceBackend ? (
          <Explorer
            category="template"
            enableBatchSync={false}
            extraColumns={[
              {
                label: 'Type',
                value: 'templateType',
                sortValue: ({ name }) => name,
              },
              {
                label: '',
                value: (t: Template) => (
                  <Button
                    id="create-resource"
                    startIcon={<Icon type={IconType.AddIcon} size="base" />}
                    onClick={event => handleAddCluster(event, t)}
                    disabled={Boolean(t.error)}
                  >
                    USE THIS TEMPLATE
                  </Button>
                ),
              },
            ]}
            linkToObject={false}
          />
        ) : (
          <DataTable
            key={templates?.length}
            filters={initialFilterState}
            rows={templates || []}
            fields={[
              {
                label: 'Name',
                value: 'name',
                sortValue: ({ name }) => name,
                textSearchable: true,
              },
              {
                label: 'Type',
                value: (t: Template) => (
                  <>{t.labels?.['weave.works/template-type']}</>
                ),
                sortValue: ({ name }) => name,
              },
              {
                label: 'Namespace',
                value: 'namespace',
                sortValue: ({ namespace }) => namespace,
              },
              {
                label: 'Provider',
                value: 'provider',
                sortValue: ({ name }) => name,
              },
              {
                label: 'Description',
                value: (t: Template) => (
                  <>
                    {t.description}
                    <Error>{t.error}</Error>
                  </>
                ),
                maxWidth: 600,
              },
              {
                label: '',
                value: (t: Template) => (
                  <Button
                    id="create-resource"
                    startIcon={<Icon type={IconType.AddIcon} size="base" />}
                    onClick={event => handleAddCluster(event, t)}
                    disabled={Boolean(t.error)}
                  >
                    USE THIS TEMPLATE
                  </Button>
                ),
              },
            ]}
            emptyMessagePlaceholder={
              <>
                <CustomEmptyMessage>
                  No templates found or no templates match the selected filter.
                  See
                </CustomEmptyMessage>
                <DocsLink
                  href="https://docs.gitops.weave.works/docs/gitops-templates/templates"
                  newTab
                >
                  here
                </DocsLink>
                <CustomEmptyMessage>
                  How to add templates and how to label them
                </CustomEmptyMessage>
              </>
            }
          />
        )}
      </NotificationsWrapper>
    </Page>
  );
};

export default TemplatesDashboard;
