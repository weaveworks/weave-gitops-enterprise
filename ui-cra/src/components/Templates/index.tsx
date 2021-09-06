import React, { FC, useMemo, useState } from 'react';
import randomColor from 'randomcolor';
import { PageTemplate } from '../Layout/PageTemplate';
import TemplateCard from './Card';
import Grid from '@material-ui/core/Grid';
import useClusters from '../../contexts/Clusters';
import useTemplates from '../../contexts/Templates';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { Loader } from '../Loader';
import { TemplatesTable } from './Table';
import styled from 'styled-components';
import { ReactComponent as GridView } from '../../assets/img/grid-view.svg';
import { ReactComponent as ListView } from '../../assets/img/list-view.svg';
import theme from 'weaveworks-ui-components/lib/theme';

const getColor = (seed: number) => {
  if (seed % 2 === 0) {
    return randomColor({
      luminosity: 'bright',
      seed: String(seed),
      hue: 'blue',
    });
  }
  return randomColor({
    luminosity: 'bright',
    seed: String(seed),
    hue: 'orange',
  });
};

const ActionsWrapper = styled.div`
  padding: ${theme.spacing.medium} ${theme.spacing.small} 0 0;
  display: flex;
  flex-direction: row;

  svg {
    width: 32px;
  }
`;

const TemplatesDashboard: FC = () => {
  const { templates, loading } = useTemplates();
  const clustersCount = useClusters().count;
  const templatesCount = templates.length;
  const [view, setView] = useState<string>('grid');

  return useMemo(() => {
    return (
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
              <ContentWrapper>
                <Grid container spacing={3} justify="center">
                  {templates.map((template: any, index: number) => (
                    <Grid key={index} item xs={11} sm={8} md={3}>
                      <TemplateCard
                        template={template}
                        color={getColor(index)}
                      />
                    </Grid>
                  ))}
                </Grid>
              </ContentWrapper>
            )}
            {view === 'table' && (
              <ContentWrapper>
                <Title>Cluster Templates</Title>
                <TemplatesTable templates={templates} />
              </ContentWrapper>
            )}
            <ActionsWrapper>
              <GridView onClick={() => setView('grid')} />
              <ListView onClick={() => setView('table')} />
            </ActionsWrapper>
          </div>
        ) : (
          <ContentWrapper>
            <Loader />
          </ContentWrapper>
        )}
      </PageTemplate>
    );
  }, [templates, loading, clustersCount, templatesCount, view]);
};

export default TemplatesDashboard;
