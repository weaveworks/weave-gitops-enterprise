import React, { FC } from 'react';
import useTemplates from '../../../contexts/Templates';
import useClusters from '../../../contexts/Clusters';
import Button from '@material-ui/core/Button';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';

const AddCluster: FC = () => {
  const { activeTemplate, setActiveTemplate } = useTemplates();
  const { addCluster, count } = useClusters();
  const handleAddCluster = (event: any) => {
    addCluster(event.target.data);
    setActiveTemplate(null);
    // redirect to Created Clusters Dashboard
  };

  return (
    <PageTemplate documentTitle="WeGo · Create new cluster">
      <span id="count-header">
        <SectionHeader
          path={[
            { label: 'Clusters', url: '/', count },
            { label: 'Create new cluster' },
          ]}
        />
      </span>
      <div>
        {activeTemplate
          ? `The active template is: ${
              activeTemplate?.name
            }. Available params in this template: ${activeTemplate.params?.map(
              p => p.name,
            )}`
          : 'Please select a template to create a cluster'}
      </div>
      <Button onClick={handleAddCluster}>Add Cluster</Button>
    </PageTemplate>
  );
};

export default AddCluster;
