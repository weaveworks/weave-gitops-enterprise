import React, { FC, useCallback, useState, useEffect } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import TemplateCard from './Card';
import Grid from '@material-ui/core/Grid';
import useNotifications, {
  NotificationData,
} from '../../contexts/Notifications';
import useTemplates from '../../hooks/templates';
import { ContentWrapper } from '../Layout/ContentWrapper';
import styled from 'styled-components';
import { ReactComponent as GridView } from '../../assets/img/grid-view.svg';
import { ReactComponent as ListView } from '../../assets/img/list-view.svg';
import {
  DataTable,
  filterConfig,
  IconType,
  theme,
  Button,
  Icon,
  Link,
} from '@weaveworks/weave-gitops';
import Autocomplete from '@material-ui/lab/Autocomplete';
import TextField from '@material-ui/core/TextField';
import { Template } from '../../cluster-services/cluster_services.pb';
import { useHistory } from 'react-router-dom';
import { TableWrapper } from '../Shared';

const ActionsWrapper = styled.div`
  display: flex;
  justify-content: end;
  svg {
    width: 32px;
    cursor: pointer;
  }
  svg.active {
    fill: ${({ theme }) => theme.colors.primary};
  }
  svg.inactive {
    fill: ${({ theme }) => theme.colors.neutral30};
  }
`;

const FilteringSection = styled.div`
  width: 100%;
  display: flex;
  justify-content: flex-end;
  padding-bottom: ${({ theme }) => theme.spacing.medium};
`;

const Error = styled.span`
  color: ${theme.colors.alert};
`;
const CustomEmptyMessage = styled.span`
  color: ${theme.colors.neutral30};
`;
const DocsLink = styled(Link)`
  color: ${theme.colors.primary};
  padding-right: ${({ theme }) => theme.spacing.xxs};
  padding-left: ${({ theme }) => theme.spacing.xxs};
`;

const TemplatesDashboard: FC<{
  location: { state: { notification: NotificationData[] } };
}> = ({ location }) => {
  const { templates, isLoading } = useTemplates();
  const { setNotifications } = useNotifications();
  const notification = location.state?.notification;
  const providers = [
    ...Array.from(new Set(templates?.map((t: Template) => t.provider))),
    'All',
  ];
  const [view, setView] = useState<string>('table');
  const [selectedProvider, setSelectedProvider] = useState<
    string | null | undefined
  >();
  const history = useHistory();

  const onProviderChange = (
    event: React.ChangeEvent<{}>,
    value: string | null | undefined,
  ) => setSelectedProvider(value);

  const validProviders = providers.filter(provider => provider !== '');

  const initialFilterState = {
    ...filterConfig(templates, 'provider'),
    ...filterConfig(templates, 'namespace'),
    ...filterConfig(templates, 'templateType'),
  };

  const filteredTemplates = selectedProvider
    ? templates?.filter(
        t => selectedProvider === 'All' || t.provider === selectedProvider,
      )
    : templates;

  const handleAddCluster = useCallback(
    (event, t) => history.push(`/templates/${t.name}/create`),
    [history],
  );

  useEffect(() => {
    if (notification) {
      setNotifications(notification);
    }
  }, [notification, setNotifications]);

  return (
    <PageTemplate
      documentTitle="Templates"
      path={[
        {
          label: 'Templates',
          url: '/templates',
        },
      ]}
    >
      <ContentWrapper loading={isLoading}>
        <div style={{ display: 'flex', flexDirection: 'column' }}>
          <ActionsWrapper id="display-action">
            <ListView
              className={view === 'table' ? 'active' : 'inactive'}
              onClick={() => setView('table')}
            />
            <GridView
              className={view === 'grid' ? 'active' : 'inactive'}
              onClick={() => setView('grid')}
            />
          </ActionsWrapper>
          {view === 'grid' && (
            <>
              <FilteringSection>
                <div style={{ width: '200px' }}>
                  <Autocomplete
                    value={selectedProvider}
                    disablePortal
                    id="filter-by-provider"
                    options={validProviders}
                    onChange={onProviderChange}
                    clearOnEscape
                    blurOnSelect="mouse"
                    renderInput={params => (
                      <TextField {...params} label="Provider" />
                    )}
                  />
                </div>
              </FilteringSection>
              <Grid container spacing={3} justifyContent="center">
                {filteredTemplates?.map((template: any, index: number) => (
                  <Grid key={index} item xs={11} sm={8} md={4}>
                    <TemplateCard template={template} />
                  </Grid>
                ))}
              </Grid>
            </>
          )}
          {view === 'table' && (
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
                        id="create-cluster"
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
          )}
        </div>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default TemplatesDashboard;
