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
import {
  CommitFile,
  RenderAutomationResponse,
  RenderTemplateResponse,
} from '../../../../cluster-services/cluster_services.pb';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { darcula } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { Button } from '@weaveworks/weave-gitops';
import JSZip from 'jszip';
import { Tooltip } from '../../../Shared';

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
interface TabContent {
  tabName: string;
  files?: CommitFile[];
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
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}
const saveAs = (content: Blob, fileName: string) => {
  const element = document.createElement('a');
  element.href = URL.createObjectURL(content);
  element.download = fileName;
  document.body.appendChild(element);
  element.click();
};

const Preview: FC<{
  openPreview: boolean;
  setOpenPreview: Dispatch<React.SetStateAction<boolean>>;
  prPreview: RenderTemplateResponse | RenderAutomationResponse;
  sourceType?: string;
  context?: string;
}> = ({ prPreview, openPreview, setOpenPreview, sourceType, context }) => {
  const initialIndex =
    context === 'app' && sourceType === 'HelmRepository' ? 1 : 0;

  const [value, setValue] = useState<number>(initialIndex);

  const handleChange = (event: React.ChangeEvent<{}>, newValue: number) => {
    setValue(newValue);
  };

  const tabsContent: Array<TabContent> =
    context === 'app'
      ? [
          {
            tabName: 'Kustomizations',
            files: prPreview.kustomizationFiles,
          },
          {
            tabName: 'Helm Releases',
            files: (prPreview as RenderAutomationResponse).helmReleaseFiles,
          },
        ]
      : [
          {
            tabName: 'Resource Definition',
            files: (prPreview as RenderTemplateResponse).renderedTemplate,
          },
          {
            tabName: 'Profiles',
            files: (prPreview as RenderTemplateResponse).profileFiles,
          },
          {
            tabName: 'Kustomizations',
            files: prPreview.kustomizationFiles,
          },
        ];

  const downloadFile = () => {
    const zip = new JSZip();
    tabsContent.forEach(tab => {
      if (tab.files) {
        tab.files.forEach(
          file => file.path && zip.file(file.path, file.content || ''),
        );
      }
    });
    zip.generateAsync({ type: 'blob' }).then(content => {
      saveAs(content, 'resources.zip');
    });
  };

  const getTooltipText = (tabName: string) => {
    switch (tabName) {
      case 'Profiles':
        return 'profile';
      case 'Helm Releases':
        return 'helm release';
      case 'Kustomizations':
        return 'kustomization';
      default:
        return '';
    }
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
        {tabsContent.map(({ tabName, files }, index) =>
          files?.length ? (
            <Tab key={index} className="tab-label" label={tabName} />
          ) : (
            <Tooltip
              title={`No ${getTooltipText(
                tabName,
              )} files in this rendered template.`}
              placement="top"
            >
              <div>
                <Tab
                  key={index}
                  className="tab-label"
                  label={tabName}
                  disabled
                />
              </div>
            </Tooltip>
          ),
        )}
      </Tabs>
      <DialogContent>
        {tabsContent.map((tab, index) => (
          <TabPanel value={value} index={index} key={index}>
            {tab.files?.map(file => (
              <div key={file.path}>
                <Typography variant="h6">{file.path}</Typography>
                <SyntaxHighlighter
                  language="yaml"
                  style={darcula}
                  wrapLongLines="pre-wrap"
                >
                  {file.content}
                </SyntaxHighlighter>
              </div>
            ))}
          </TabPanel>
        ))}
      </DialogContent>
      <div className="info">
        <span>
          You may edit these as part of the pull request with your git provider.
        </span>
        <Button onClick={downloadFile}>Download</Button>
      </div>
    </DialogWrapper>
  );
};

export default Preview;
