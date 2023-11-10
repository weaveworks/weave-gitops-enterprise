import {
  Button,
  DataTable,
  filterConfig,
  Icon,
  IconType,
  Link,
} from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { FC, useCallback, useEffect } from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { EnabledComponent, Object } from '../../api/query/query.pb';
import { Template } from '../../cluster-services/cluster_services.pb';
import useNotifications, {
  NotificationData,
} from '../../contexts/Notifications';
import { useIsEnabledForComponent } from '../../hooks/query';
import useTemplates from '../../hooks/templates';
import Explorer from '../Explorer/Explorer';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';

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

type Props = {
  className?: string;
  location: { state: { notification: NotificationData[] } };
};

const TemplatesDashboard: FC<Props> = ({ location, className }) => {
  const isExplorerEnabled = useIsEnabledForComponent(
    EnabledComponent.templates,
  );
  const { templates, isLoading } = useTemplates({
    enabled: !isExplorerEnabled,
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
      className={className}
      path={[
        {
          label: 'Templates',
        },
      ]}
    >
      <NotificationsWrapper>
        {isExplorerEnabled ? (
          <Explorer
            category="template"
            enableBatchSync={false}
            extraColumns={[
              {
                label: 'Type',
                value: o => {
                  return _.get(o.parsed, [
                    'metadata',
                    'labels',
                    'weave.works/template-type',
                  ]);
                },
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
              {
                label: 'Description',
                value: (o: Object) => o.message || '-',
                index: 7,
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

export default styled(TemplatesDashboard)`
  ${Explorer} {
    /* Hiding Status, Message, and Tenant columns */
    table td:nth-child(5),
    table th:nth-child(5),
    table td:nth-child(6),
    table th:nth-child(6),
    table td:nth-child(7),
    table th:nth-child(7) {
      display: none;
    }
  }
`;
