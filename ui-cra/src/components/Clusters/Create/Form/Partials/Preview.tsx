import React, { FC, Dispatch } from 'react';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import { CloseIconButton } from '../../../../../assets/img/close-icon-button';
import {
  Tab,
  Tabs,
  Typography,
  DialogContent,
  DialogTitle,
  Dialog,
  TextareaAutosize,
  Box,
} from '@material-ui/core';

const useStyles = makeStyles(() =>
  createStyles({
    textarea: {
      width: '100%',
      padding: weaveTheme.spacing.xs,
      border: `1px solid ${weaveTheme.colors.neutral20}`,
    },
    info: {
      padding: weaveTheme.spacing.medium,
    },
    tabsContainer: {
      marginLeft: weaveTheme.spacing.large,
    },
    tabLabel: {
      color: weaveTheme.colors.primary10,
      fontSize: weaveTheme.fontSizes.small,
    },
  }),
);

const Preview: FC<{
  openPreview: boolean;
  setOpenPreview: Dispatch<React.SetStateAction<boolean>>;
  PRPreview: {
    renderedTemplate: string;
    kustomizationFiles: { path: string; content: string }[];
    profileFiles: { path: string; content: string }[];
  };
}> = ({ PRPreview, openPreview, setOpenPreview }) => {
  const classes = useStyles();
  const [value, setValue] = React.useState(0);

  const handleChange = (event: React.ChangeEvent<{}>, newValue: number) => {
    setValue(newValue);
  };

  interface TabPanelProps {
    children?: React.ReactNode;
    index: number;
    value: number;
  }

  function TabPanel(props: TabPanelProps) {
    const { children, value, index, ...other } = props;

    return (
      <div
        role="tabpanel"
        hidden={value !== index}
        id={`tabpanel-${index}`}
        aria-labelledby={`tab-${index}`}
        {...other}
      >
        {value === index && (
          <Box sx={{ p: 3 }}>
            <Typography>{children}</Typography>
          </Box>
        )}
      </div>
    );
  }

  return (
    <Dialog
      open={openPreview}
      maxWidth="md"
      fullWidth
      scroll="paper"
      onClose={() => setOpenPreview(false)}
    >
      <DialogTitle disableTypography>
        <Typography variant="h5">PR Preview</Typography>
        <CloseIconButton onClick={() => setOpenPreview(false)} />
      </DialogTitle>
      <Tabs
        className={classes.tabsContainer}
        indicatorColor="primary"
        value={value}
        onChange={handleChange}
        aria-label="pr-preview-sections"
      >
        {['Cluster Definition', 'Profiles', 'Kustomizations'].map(
          (tabName, index) => (
            <Tab key={index} className={classes.tabLabel} label={tabName} />
          ),
        )}
      </Tabs>
      <DialogContent>
        <TabPanel value={value} index={0}>
          <TextareaAutosize
            className={classes.textarea}
            value={PRPreview.renderedTemplate}
            readOnly
          />
        </TabPanel>
        <TabPanel value={value} index={1}>
          <TextareaAutosize
            className={classes.textarea}
            value={PRPreview.profileFiles.map(
              profileFile => profileFile.content + '---',
            )}
            readOnly
          />
        </TabPanel>
        <TabPanel value={value} index={2}>
          <TextareaAutosize
            className={classes.textarea}
            value={PRPreview.kustomizationFiles.map(
              kustomizationFile => kustomizationFile.content + '---',
            )}
            readOnly
          />
        </TabPanel>
      </DialogContent>
      <div className={classes.info}>
        <span>
          You may edit these as part of the pull request with your git provider.
        </span>
      </div>
    </Dialog>
  );
};

export default Preview;
