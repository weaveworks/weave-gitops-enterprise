import React, { FC, useCallback, useState } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import TemplateCard from './Card';
import Grid from '@material-ui/core/Grid';
import useClusters from '../../contexts/Clusters';
import useTemplates from '../../contexts/Templates';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import styled from 'styled-components';
import { ReactComponent as GridView } from '../../assets/img/grid-view.svg';
import { ReactComponent as ListView } from '../../assets/img/list-view.svg';
import {
  FilterableTable,
  filterConfig,
  IconType,
  theme,
  Button,
  Icon,
} from '@weaveworks/weave-gitops';
import Autocomplete from '@material-ui/lab/Autocomplete';
import TextField from '@material-ui/core/TextField';
import { ThemeProvider, createTheme } from '@material-ui/core/styles';
import { muiTheme } from '../../muiTheme';
import { Template } from '../../cluster-services/cluster_services.pb';
import { useHistory } from 'react-router-dom';
import { TableWrapper } from '../Shared';

const ActionsWrapper = styled.div`
  padding: ${({ theme }) => theme.spacing.small} 0
    ${({ theme }) => theme.spacing.medium} 0;
  display: flex;
  justify-content: end;
  svg {
    width: 32px;
  }
  svg.active {
    fill: ${({ theme }) => theme.colors.primary};
    cursor: pointer;
  }
  svg.inactive {
    fill: ${({ theme }) => theme.colors.neutral30};
    cursor: pointer;
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

const localMuiTheme = createTheme({
  ...muiTheme,
  overrides: {
    ...muiTheme.overrides,
    MuiInputBase: {
      ...muiTheme.overrides?.MuiInputBase,
      input: {
        ...muiTheme.overrides?.MuiInputBase?.input,
        border: 'none',
        position: 'static',
        backgroundColor: 'transparent',
      },
    },
  },
});

const TemplatesDashboard: FC = () => {
  const { templates, isLoading } = useTemplates();
  const providers = [
    ...Array.from(new Set(templates?.map((t: Template) => t.provider))),
    'All',
  ];
  const clustersCount = useClusters().count;
  const templatesCount = templates?.length;
  const [view, setView] = useState<string>('grid');
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
    ...filterConfig(templates, 'templateKind'),
  };

  const filteredTemplates = selectedProvider
    ? templates?.filter(
        t => selectedProvider === 'All' || t.provider === selectedProvider,
      )
    : templates;

  const handleAddCluster = useCallback(
    (event, t) => history.push(`/clusters/templates/${t.name}/create`),
    [history],
  );

  return (
    <ThemeProvider theme={localMuiTheme}>
      <PageTemplate documentTitle="WeGO · Templates">
        <SectionHeader
          path={[
            { label: 'Clusters', url: '/clusters', count: clustersCount },
            {
              label: 'Templates',
              url: '/clusters/templates',
              count: templatesCount,
            },
          ]}
        />

        {/* backgroundColor={view === 'grid' ? 'transparent' : undefined} */}
        <ContentWrapper loading={isLoading}>
          <div style={{ display: 'flex', flexDirection: 'column' }}>
            <ActionsWrapper>
              <GridView
                className={view === 'grid' ? 'active' : 'inactive'}
                onClick={() => setView('grid')}
              />
              <ListView
                className={view === 'table' ? 'active' : 'inactive'}
                onClick={() => setView('table')}
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
                <FilterableTable
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
                      label: 'Kind',
                      value: 'templateKind',
                      sortValue: ({ name }) => name,
                      textSearchable: true,
                    },
                    {
                      label: 'Provider',
                      value: 'provider',
                      sortValue: ({ name }) => name,
                      textSearchable: true,
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
                          startIcon={
                            <Icon type={IconType.AddIcon} size="base" />
                          }
                          onClick={event => handleAddCluster(event, t)}
                          disabled={Boolean(t.error)}
                        >
                          CREATE CLUSTER WITH THIS TEMPLATE
                        </Button>
                      ),
                    },
                  ]}
                />
              </TableWrapper>
            )}
          </div>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default TemplatesDashboard;
