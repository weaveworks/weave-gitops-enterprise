import React, { FC, Dispatch, useState } from 'react';
import styled from 'styled-components';
import { CloseIconButton } from '../../../../assets/img/close-icon-button';
import {
  Tab,
  Tabs,
  Typography,
  DialogContent,
  DialogTitle,
  Dialog,
  Box,
} from '@material-ui/core';
import { AppPRPreview, ClusterPRPreview } from '../../../../types/custom';
import {
  CommitFile,
  RenderTemplateResponse,
} from '../../../../cluster-services/cluster_services.pb';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { darcula } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { Button } from '@weaveworks/weave-gitops';

const DialogWrapper = styled(Dialog)`
  div[class*='MuiPaper-root'] {
    height: 700px;
  }
  .info {
    padding: ${({ theme }) => theme.spacing.medium};
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .tabs-container {
    margin-left: ${({ theme }) => theme.spacing.large};
  }
  .tab-label {
    color: ${({ theme }) => theme.colors.primary10};
    font-size: ${({ theme }) => theme.fontSizes.small};
  }
`;
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

const Preview: FC<{
  openPreview: boolean;
  setOpenPreview: Dispatch<React.SetStateAction<boolean>>;
  PRPreview: RenderTemplateResponse;
  sourceType?: string;
  context?: string;
}> = ({ PRPreview, openPreview, setOpenPreview, sourceType, context }) => {
  const [value, setValue] = useState<number>(0);

  const handleChange = (event: React.ChangeEvent<{}>, newValue: number) => {
    setValue(newValue);
  };

  const getContetn = (files: CommitFile[] | undefined) =>
    files?.map(file => file.content).join('\n---\n');

  const tabsContent =
    context === 'app'
      ? [
          {
            tabName: 'Kustomizations',
            value: getContetn(PRPreview.kustomizationFiles),
            fileGroup: 'resources',
          },
          {
            tabName: 'Helm Releases',
            value: getContetn((PRPreview as AppPRPreview).helmReleaseFiles),
            fileGroup: 'resources',
          },
        ]
      : [
          {
            tabName: 'Cluster Definition',
            value: (PRPreview as ClusterPRPreview).renderedTemplate,
            fileGroup: 'cluster_definition',
          },
          {
            tabName: 'Profiles',
            value: getContetn((PRPreview as ClusterPRPreview).profileFiles),
            fileGroup: 'resources',
          },
          {
            tabName: 'Kustomizations',
            value: getContetn(PRPreview.kustomizationFiles),
            fileGroup: 'resources',
          },
        ];

  const downloadFile = () => {
    let files: { [key: string]: BlobPart } = {};
    tabsContent.forEach(tab => {
      if (!files[tab.fileGroup]) {
        files[tab.fileGroup] = tab.value as BlobPart;
      } else {
        const content = files[tab.fileGroup] + '\n---\n' + tab.value;
        files[tab.fileGroup] = content as BlobPart;
      }
    });

    Object.entries(files).forEach(([key, value]) => {
      const file = new Blob([value], { type: 'yaml' });
      const element = document.createElement('a');
      element.href = URL.createObjectURL(file);
      element.download = `${key}.yaml`;
      document.body.appendChild(element);
      element.click();
    });
  };
  return (
    <DialogWrapper
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
        className="tabs-container"
        indicatorColor="primary"
        value={value}
        onChange={handleChange}
        aria-label="pr-preview-sections"
        selectionFollowsFocus={true}
      >
        {tabsContent.map(({ tabName }, index) => (
          <Tab key={index} className="tab-label" label={tabName} />
        ))}
      </Tabs>
      <DialogContent>
        {tabsContent.map((tab, index) => (
          <TabPanel value={value} index={index} key={index}>
            <SyntaxHighlighter
              language="yaml"
              style={darcula}
              wrapLongLines="pre-wrap"
              showLineNumbers={true}
            >
              {tab.value}
            </SyntaxHighlighter>
          </TabPanel>
        ))}
      </DialogContent>
      <div className="info" >
        <span>
          You may edit these as part of the pull request with your git provider.
        </span>
        <Button onClick={downloadFile}>Download</Button>
      </div>
    </DialogWrapper>
  );
};

export default Preview;
