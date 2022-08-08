import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import {
  AutomationsTable,
  Button,
  Icon,
  IconType,
  LoadingPage,
  useListAutomations,
} from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { useHistory } from 'react-router-dom';
import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';

interface Size {
  size?: 'small';
}
const ActionsWrapper = styled.div<Size>`
  display: flex;
  & > * {
    margin-right: ${({ theme }) => theme.spacing.medium};
  }
`;

const WGApplicationsDashboard: FC = () => {
  const { data: automations, isLoading } = useListAutomations();
  const applicationsCount = useApplicationsCount();
  const history = useHistory();

  const handleAddApplication = () => {
    history.push('/applications/new');
  };

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGO · Applications">
        <SectionHeader
          path={[
            {
              label: 'Applications',
              url: '/applications',
              count: applicationsCount,
            },
          ]}
        />
        <ContentWrapper errors={automations?.errors}>
          <ActionsWrapper>
            <Button
              id="add-application"
              startIcon={<Icon type={IconType.AddIcon} size="base" />}
              onClick={handleAddApplication}
            >
              ADD AN APPLICATION
            </Button>
          </ActionsWrapper>
          {isLoading ? (
            <LoadingPage />
          ) : (
            <AutomationsTable automations={automations?.result} />
          )}
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default WGApplicationsDashboard;
