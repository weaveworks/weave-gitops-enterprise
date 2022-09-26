import React, { FC, Dispatch, useState, useEffect } from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { CloseIconButton } from '../../../../assets/img/close-icon-button';
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
import { AppPRPreview, ClusterPRPreview } from '../../../../types/custom';

const DialogWrapper = styled(Dialog)`
  div[class*='MuiPaper-root']{
    height: 700px;
  }
  .textarea {
    width: 100%;
      padding:  ${({ theme }) => theme.spacing.xs};
      border: 1px solid ${({ theme }) => theme.colors.neutral20};
    }
  }
  .info {
    padding:  ${({ theme }) => theme.spacing.medium};
  }
  .tabs-container{
    margin-left:  ${({ theme }) => theme.spacing.large};
  }
  .tab-label {
    color: ${({ theme }) => theme.colors.primary10};
    font-size: ${({ theme }) => theme.fontSizes.small}
  }
`;

const Preview: FC<{
  openPreview: boolean;
  setOpenPreview: Dispatch<React.SetStateAction<boolean>>;
  PRPreview: ClusterPRPreview | AppPRPreview;
  context?: string;
}> = ({ PRPreview, openPreview, setOpenPreview, context }) => {
  const history = useHistory();
  const [value, setValue] = useState<number>(0);

  const handleChange = (event: React.ChangeEvent<{}>, newValue: number) => {
    setValue(newValue);
  };

  interface TabPanelProps {
    children?: React.ReactNode;
    index: number;
    value: number;
    empty?: boolean;
  }

  function TabPanel(props: TabPanelProps) {
    const { children, value, index, empty, ...other } = props;

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

  useEffect(() => {
    if (context === 'app' && PRPreview.kustomizationFiles.length === 0) {
      console.log('setting it');
      setValue(1);
    }
    // return history.listen(() => setValue(0));
  }, [PRPreview.kustomizationFiles.length, context, history]);

  console.log(PRPreview.kustomizationFiles.length);

  return (
    <DialogWrapper
      open={openPreview}
      maxWidth="md"
      fullWidth
      scroll="paper"
      onClose={() => {
        setOpenPreview(false);
        setValue(0);
      }}
    >
      <DialogTitle disableTypography>
        <Typography variant="h5">PR Preview</Typography>
        <CloseIconButton onClick={() => setOpenPreview(false)} />
      </DialogTitle>
      <Tabs
        className="tabs-container"
        indicatorColor="primary"
        value={value}
        onChange={handleChange}
        aria-label="pr-preview-sections"
        selectionFollowsFocus={true}
      >
        {(context === 'app'
          ? ['Kustomizations', 'Helm Releases']
          : ['Cluster Definition', 'Profiles', 'Kustomizations']
        ).map((tabName, index) => (
          <Tab key={index} className="tab-label" label={tabName} />
        ))}
      </Tabs>
      <DialogContent>
        {context !== 'app' ? (
          <>
            <TabPanel value={value} index={0}>
              <TextareaAutosize
                className="textarea"
                value={(PRPreview as ClusterPRPreview).renderedTemplate}
                readOnly
              />
            </TabPanel>
            <TabPanel value={value} index={1}>
              <TextareaAutosize
                className="textarea"
                value={(PRPreview as ClusterPRPreview).profileFiles
                  ?.map(profileFile => profileFile.content)
                  .join('\n---\n')}
                readOnly
              />
            </TabPanel>
            <TabPanel value={value} index={2}>
              <TextareaAutosize
                className="textarea"
                value={PRPreview.kustomizationFiles
                  ?.map(kustomizationFile => kustomizationFile.content)
                  .join('\n---\n')}
                readOnly
              />
            </TabPanel>
          </>
        ) : (
          <>
            <TabPanel value={value} index={0}>
              <TextareaAutosize
                className="textarea"
                value={PRPreview.kustomizationFiles
                  ?.map(kustomizationFile => kustomizationFile.content)
                  .join('\n---\n')}
                readOnly
              />
            </TabPanel>
            <TabPanel value={value} index={1}>
              <TextareaAutosize
                className="textarea"
                value={(PRPreview as AppPRPreview).helmReleaseFiles
                  ?.map(helmReleaseFile => helmReleaseFile.content)
                  .join('\n---\n')}
                readOnly
              />
            </TabPanel>
          </>
        )}
      </DialogContent>
      <div className="info">
        <span>
          You may edit these as part of the pull request with your git provider.
        </span>
      </div>
    </DialogWrapper>
  );
};

export default Preview;
