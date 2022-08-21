import { Grid } from '@material-ui/core';
import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import React, { Dispatch, FC } from 'react';
import AppFields from '../../../../Applications/Add/form/Partials/AppFields';

export const ApplicationsWrapper: FC<{
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
}> = ({ formData, setFormData }) => {
  const handleAddApplication = () => {
    let newAutomations = [...formData.clusterAutomations];
    newAutomations.push({ name: '', namespace: '', path: '' });
    setFormData({ ...formData, clusterAutomations: newAutomations });
  };
  const handleRemoveApplication = (index: number) => {
    let newAutomations = [...formData.clusterAutomations];
    newAutomations.splice(index, 1);
    setFormData({ ...formData, clusterAutomations: newAutomations });
  };

  return (
    <>
      <h2>Applications</h2>
      {formData.clusterAutomations.map((kustomization: any, index: number) => {
        return (
          <div key={index}>
            <Grid container className="">
              <Grid item xs={12} sm={8} md={8} lg={8}>
                <AppFields
                  formData={formData}
                  setFormData={setFormData}
                  index={index}
                  isMultiple={true}
                ></AppFields>
              </Grid>
              <Grid
                item
                xs={12}
                sm={4}
                md={4}
                lg={4}
                justifyContent="center"
                alignItems="center"
                container
              >
                <Button
                  id="remove-application"
                  startIcon={<Icon type={IconType.DeleteIcon} size="base" />}
                  onClick={() => handleRemoveApplication(index)}
                >
                  REMOVE APPLICATION
                </Button>
              </Grid>
            </Grid>
          </div>
        );
      })}
      <Button
        id="add-application"
        startIcon={<Icon type={IconType.AddIcon} size="base" />}
        onClick={handleAddApplication}
      >
        ADD AN APPLICATION
      </Button>
    </>
  );
};
