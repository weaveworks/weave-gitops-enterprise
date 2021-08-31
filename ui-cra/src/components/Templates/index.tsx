import React, { FC } from 'react';
import randomColor from 'randomcolor';
import { PageTemplate } from '../Layout/PageTemplate';
import TemplateCard from './Card';
import Grid from '@material-ui/core/Grid';
import useClusters from '../../contexts/Clusters';
import useTemplates from '../../contexts/Templates';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { Loader } from '../Loader';

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

const TemplatesDashboard: FC = () => {
  const { templates, loading } = useTemplates();
  const clustersCount = useClusters().count;
  const templatesCount = templates.length;

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
      <ContentWrapper>
        {!loading ? (
          <Grid container spacing={3} justify="center">
            {templates.map((template: any, index: number) => (
              <Grid key={index} item xs={12} sm={9} md={4}>
                <TemplateCard template={template} color={getColor(index)} />
              </Grid>
            ))}
          </Grid>
        ) : (
          <Loader />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default TemplatesDashboard;
