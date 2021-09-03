import React, { FC, useState } from 'react';
import randomColor from 'randomcolor';
import { PageTemplate } from '../Layout/PageTemplate';
import TemplateCard from './Card';
import Grid from '@material-ui/core/Grid';
import useClusters from '../../contexts/Clusters';
import useTemplates from '../../contexts/Templates';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { Loader } from '../Loader';
import { TemplatesTable } from './Table';
import styled from 'styled-components';
import { ReactComponent as GridView } from '../../assets/img/grid-view.svg';
import { ReactComponent as ListView } from '../../assets/img/list-view.svg';
import theme from 'weaveworks-ui-components/lib/theme';
import { Button } from '@material-ui/core';

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
  min-width: 36px;
  padding: ${theme.spacing.medium} ${theme.spacing.small} 0
    ${theme.spacing.small};
`;

const TemplatesDashboard: FC = () => {
  const { templates, loading } = useTemplates();
  const clustersCount = useClusters().count;
  const templatesCount = templates.length;
  const [view, setView] = useState<string>('grid');

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
          <ContentWrapper>
            {view === 'grid' && (
              <Grid container spacing={3} justify="center">
                {templates.map((template: any, index: number) => (
                  <Grid key={index} item xs={11} sm={8} md={3}>
                    <TemplateCard template={template} color={getColor(index)} />
                  </Grid>
                ))}
              </Grid>
            )}
            {view === 'table' && <TemplatesTable templates={templates} />}
          </ContentWrapper>
          <ActionsWrapper>
            <Button
              onClick={() => setView('grid')}
              variant="contained"
              endIcon={<GridView style={{ width: '50px' }} />}
            />
            <Button
              onClick={() => setView('table')}
              variant="contained"
              endIcon={<ListView style={{ width: '50px' }} />}
            />
          </ActionsWrapper>
        </div>
      ) : (
        <ContentWrapper>
          <Loader />
        </ContentWrapper>
      )}
    </PageTemplate>
  );
};

export default TemplatesDashboard;
