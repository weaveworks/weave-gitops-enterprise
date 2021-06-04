import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import TemplateCard from './Card';
import Grid from '@material-ui/core/Grid';
import useClusters from '../../contexts/Clusters';
import useTemplates from '../../contexts/Templates';
import { SectionHeader } from '../Layout/SectionHeader';

const TemplatesDashboard: FC = () => {
  const { templates } = useTemplates();
  const clustersCount = useClusters().count;
  const templatesCount = templates.length;

  return (
    <PageTemplate documentTitle="WeGO · Templates">
      <span id="count-header">
        <SectionHeader
          path={[
            { label: 'Clusters', url: 'clusters', count: clustersCount },
            { label: 'Templates', url: 'templates', count: templatesCount },
          ]}
        />
      </span>
      <Grid container spacing={4} justify="center">
        {templates.map((template: any, index: number) => (
          <Grid key={index} item xs={12} sm={6} md={4}>
            <TemplateCard template={template} />
          </Grid>
        ))}
      </Grid>
    </PageTemplate>
  );
};

export default TemplatesDashboard;
