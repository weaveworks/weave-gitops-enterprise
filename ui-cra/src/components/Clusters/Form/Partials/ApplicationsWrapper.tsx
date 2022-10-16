import { Grid, makeStyles, createStyles } from '@material-ui/core';
import { Button, Icon, IconType, theme } from '@weaveworks/weave-gitops';
import React, { Dispatch, FC } from 'react';
import AppFields from '../../../Applications/Add/form/Partials/AppFields';

const useStyles = makeStyles(() =>
  createStyles({
    removeApplicationWrapper: {
      paddingRight: theme.spacing.base,
    },
    addApplicationSectionWrapper: {
      paddingBottom: theme.spacing.xl,
    },
    applicationWrapper: {
      border: `1px solid ${theme.colors.neutral20}`,
      padding: theme.spacing.small,
      marginBottom: theme.spacing.small,
      borderRadius: theme.spacing.xxs,
      borderStyle: 'dashed',
    },
  }),
);

export const ApplicationsWrapper: FC<{
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  isKustomizationsEnabled?: string;
}> = ({ formData, setFormData, isKustomizationsEnabled = 'true' }) => {
  const classes = useStyles();

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

  return isKustomizationsEnabled === 'false' ? null : (
    <div className={classes.addApplicationSectionWrapper}>
      <h2>Applications</h2>
      {formData.clusterAutomations?.map((kustomization: any, index: number) => {
        return (
          <div key={index} className={classes.applicationWrapper}>
            <h3>Application No.{index + 1}</h3>
            <Grid container>
              <Grid item xs={12} sm={8} md={8} lg={8}>
                <AppFields
                  index={index}
                  formData={formData}
                  setFormData={setFormData}
                  allowSelectCluster={false}
                />
              </Grid>
              <Grid
                item
                xs={12}
                sm={4}
                md={4}
                lg={4}
                justifyContent="flex-end"
                alignItems="center"
                container
              >
                <div className={classes.removeApplicationWrapper}>
                  <Button
                    id="remove-application"
                    startIcon={<Icon type={IconType.DeleteIcon} size="base" />}
                    onClick={() => handleRemoveApplication(index)}
                  >
                    REMOVE APPLICATION
                  </Button>
                </div>
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
    </div>
  );
};
