import Grid from '@material-ui/core/Grid';
import { createTheme, ThemeProvider } from '@material-ui/core/styles';
import TextField from '@material-ui/core/TextField';
import Autocomplete from '@material-ui/lab/Autocomplete';
import { theme } from '@weaveworks/weave-gitops';
import React, { FC, useState } from 'react';
import styled from 'styled-components';
import { ReactComponent as GridView } from '../../assets/img/grid-view.svg';
import { ReactComponent as ListView } from '../../assets/img/list-view.svg';
import useClusters from '../../contexts/Clusters';
import useTemplates from '../../contexts/Templates';
import { muiTheme } from '../../muiTheme';
import { Template } from '../../types/custom';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { Loader } from '../Loader';
import TemplateCard from './Card';
import { TemplatesTable } from './Table';

const ActionsWrapper = styled.div`
  padding: ${theme.spacing.medium} ${theme.spacing.small} 0 0;
  display: flex;

  svg {
    width: 32px;
  }

  svg.inactive {
    fill: ${theme.colors.neutral20};
  }
`;

const TitleSection = styled.div`
  width: 100%;
  display: flex;
  justify-content: space-between;
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
  const { templates, loading } = useTemplates();
  const providers = [
    ...Array.from(new Set(templates.map((t: Template) => t.provider))),
    'All',
  ];
  const clustersCount = useClusters().count;
  const templatesCount = templates.length;
  const [view, setView] = useState<string>('grid');
  const [selectedProvider, setSelectedProvider] = useState<
    string | null | undefined
  >();

  const onProviderChange = (
    event: React.ChangeEvent<{}>,
    value: string | null | undefined,
  ) => setSelectedProvider(value);

  const validProviders = providers.filter(provider => provider !== '');

  const titleSection = (
    <TitleSection>
      {view === 'grid' ? (
        <Title style={{ paddingBottom: theme.spacing.xl }}>
          Cluster Templates
        </Title>
      ) : (
        <Title>Cluster Templates</Title>
      )}
      <div style={{ width: '200px' }}>
        <Autocomplete
          value={selectedProvider}
          disablePortal
          id="filter-by-provider"
          options={validProviders}
          onChange={onProviderChange}
          clearOnEscape
          blurOnSelect="mouse"
          renderInput={params => <TextField {...params} label="Provider" />}
        />
      </div>
    </TitleSection>
  );

  const filteredTemplates = selectedProvider
    ? templates.filter(
        t => selectedProvider === 'All' || t.provider === selectedProvider,
      )
    : templates;

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
        {!loading ? (
          <div style={{ display: 'flex' }}>
            {view === 'grid' && (
              <ContentWrapper backgroundColor="transparent">
                {titleSection}
                <Grid container spacing={3} justifyContent="center">
                  {filteredTemplates.map((template: any, index: number) => (
                    <Grid key={index} item xs={11} sm={8} md={4}>
                      <TemplateCard template={template} />
                    </Grid>
                  ))}
                </Grid>
              </ContentWrapper>
            )}
            {view === 'table' && (
              <ContentWrapper>
                {titleSection}
                <TemplatesTable
                  key={filteredTemplates.length}
                  templates={filteredTemplates}
                />
              </ContentWrapper>
            )}
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
          </div>
        ) : (
          <ContentWrapper>
            <Loader />
          </ContentWrapper>
        )}
      </PageTemplate>
    </ThemeProvider>
  );
};

export default TemplatesDashboard;
