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
          },
          {
            tabName: 'Helm Releases',
            value: getContetn((PRPreview as AppPRPreview).helmReleaseFiles),
          },
        ]
      : [
          {
            tabName: 'Cluster Definition',
            value: (PRPreview as ClusterPRPreview).renderedTemplate,
          },
          {
            tabName: 'Profiles',
            value: getContetn((PRPreview as ClusterPRPreview).profileFiles),
          },
          {
            tabName: 'Kustomizations',
            value: getContetn(PRPreview.kustomizationFiles),
          },
        ];

  tabsContent.push({
    tabName: 'Full',
    value: tabsContent.reduce((prev, next) => {
      return next.value ? prev + next.value + '\n---\n' : prev;
    }, ''),
  });

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
      <div className="info">
        <span>
          You may edit these as part of the pull request with your git provider.
        </span>
      </div>
    </DialogWrapper>
  );
};

export default Preview;
