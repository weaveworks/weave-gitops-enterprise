import React, { FC, useEffect, useState } from 'react';
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
  applicationsClient,
  theme,
} from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { useHistory } from 'react-router-dom';
import { useListConfig } from '../../hooks/versions';

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
  const { data } = useListConfig();
  const repositoryURL = data?.repositoryURL || '';
  const [repoLink, setRepoLink] = useState<string>('');

  const handleAddApplication = () => {
    history.push('/applications/create');
  };

  useEffect(() => {
    repositoryURL &&
      applicationsClient.ParseRepoURL({ url: repositoryURL }).then(res => {
        if (res.provider === 'GitHub') {
          setRepoLink(repositoryURL + `/pulls`);
        } else if (res.provider === 'GitLab') {
          setRepoLink(repositoryURL + `/-/merge_requests`);
        }
      });
  }, [repositoryURL]);

  return (
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
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <ActionsWrapper>
            <Button
              id="add-application"
              startIcon={<Icon type={IconType.AddIcon} size="base" />}
              onClick={handleAddApplication}
            >
              ADD AN APPLICATION
            </Button>
          </ActionsWrapper>
          <a
            style={{
              color: theme.colors.primary,
              padding: theme.spacing.small,
            }}
            href={repoLink}
            target="_blank"
            rel="noopener noreferrer"
          >
            View open Pull Requests
          </a>
        </div>
        {isLoading ? (
          <LoadingPage />
        ) : (
          <AutomationsTable automations={automations?.result} />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsDashboard;
