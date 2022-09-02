import { Grid, makeStyles, createStyles } from '@material-ui/core';
import { Button, Icon, IconType, theme } from '@weaveworks/weave-gitops';
import React, { Dispatch, FC, useCallback, useState } from 'react';
import AppFields from '../../../../Applications/Add/form/Partials/AppFields';
import useTemplates from '../../../../../contexts/Templates';
import useNotifications from '../../../../../contexts/Notifications';
import Preview from './Preview';

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
}> = ({ formData, setFormData }) => {
  const {
    renderTemplate,
  } = useTemplates();
  const [openPreview, setOpenPreview] = useState(false);
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [PRPreview, setPRPreview] = useState<string | null>(null);
  const { setNotifications } = useNotifications();
  const classes = useStyles();

  const handlePRPreview = useCallback(() => {
    const { ...templateFields } = formData;
    setPreviewLoading(true);
    return renderTemplate({
      values: templateFields,
    })
      .then(data => {
        setOpenPreview(true);
        setPRPreview(data.renderedTemplate);
      })
      .catch(err =>
        setNotifications([
          { message: { text: err.message }, variant: 'danger' },
        ]),
      )
      .finally(() => setPreviewLoading(false));
  }, [
    formData,
    setOpenPreview,
    renderTemplate,
    setNotifications,
  ]);


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
    <div className={classes.addApplicationSectionWrapper}>
      <h2>Applications</h2>
      {formData.clusterAutomations.map((kustomization: any, index: number) => {
        return (
          <div key={index} className={classes.applicationWrapper}>
            <h3>Application No.{index + 1}</h3>
            <Grid container className="">
              <Grid item xs={12} sm={8} md={8} lg={8}>
                <AppFields
                  index={index}
                  formData={formData}
                  setFormData={setFormData}
                  onPRPreview={handlePRPreview}
                  previewLoading={previewLoading}
                />
              </Grid>
              {openPreview && PRPreview ? (
                <Preview
                  openPreview={openPreview}
                  setOpenPreview={setOpenPreview}
                  PRPreview={PRPreview}
                />
              ) : null}
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
