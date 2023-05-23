import React, { FC, useCallback, useEffect } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import useTemplates from '../../hooks/templates';
import useNotifications, {
  NotificationData,
} from '../../contexts/Notifications';
import { ContentWrapper } from '../Layout/ContentWrapper';
import styled from 'styled-components';

import {
  DataTable,
  filterConfig,
  IconType,
  theme,
  Button,
  Icon,
  Link,
} from '@weaveworks/weave-gitops';
import { Template } from '../../cluster-services/cluster_services.pb';
import { useLocation, useNavigate } from 'react-router-dom';
import { TableWrapper } from '../Shared';

const Error = styled.span`
  color: ${theme.colors.alertOriginal};
`;
const CustomEmptyMessage = styled.span`
  color: ${theme.colors.neutral30};
`;
const DocsLink = styled(Link)`
  color: ${theme.colors.primary};
  padding-right: ${({ theme }) => theme.spacing.xxs};
  padding-left: ${({ theme }) => theme.spacing.xxs};
`;

const TemplatesDashboard: FC<{}> = () => {
  const location = useLocation();
  const { templates, isLoading } = useTemplates();
  const { setNotifications } = useNotifications();
  const navigate = useNavigate();

  const initialFilterState = {
    ...filterConfig(templates, 'provider'),
    ...filterConfig(templates, 'namespace'),
    ...filterConfig(templates, 'templateType'),
  };

  const handleAddCluster = useCallback(
    (event, t) => navigate(`/templates/${t.name}/create`),
    [navigate],
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
    <PageTemplate
      documentTitle="Templates"
      path={[
        {
          label: 'Templates',
        },
      ]}
    >
      <ContentWrapper loading={isLoading}>
        <div style={{ display: 'flex', flexDirection: 'column' }}>
          <TableWrapper id="templates-list">
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
                  value: 'templateType',
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
                    No templates found or no templates match the selected
                    filter. See
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
          </TableWrapper>
        </div>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default TemplatesDashboard;
