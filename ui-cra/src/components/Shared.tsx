import React, { AnchorHTMLAttributes, FC, Key } from 'react';
import styled from 'styled-components';
import { theme } from '@weaveworks/weave-gitops';
import { sortBy } from 'lodash';
import { Tooltip as Mtooltip, TooltipProps } from '@material-ui/core';
import {
  Cluster,
  ClusterStatus,
  GitCommitInfo,
  FluxInfo,
} from '../types/kubernetes';
import GitUrlParse from 'git-url-parse';
import { SparkTimeline } from './SparkTimeline';
import { getClusterStatus, ReadyStatus } from './Clusters/Status';
import { Loader } from './Loader';
import Box from '@material-ui/core/Box';

export const SafeAnchor: FC<AnchorHTMLAttributes<HTMLAnchorElement>> = ({
  children,
  className,
  title,
  href,
}) => (
  <a
    className={className}
    title={title}
    href={href}
    rel="noopener noreferrer"
    target="_blank"
  >
    {children}
  </a>
);

export interface FinishMessage {
  success: boolean;
  message: string;
}

export const NameLink = styled(SafeAnchor)`
  white-space: nowrap;
  color: ${theme.colors.primary};
  &:hover {
    color: ${theme.colors.primaryDark};
  }
`;

const Name = styled.span`
  white-space: nowrap;
`;

export const ClusterNameLink: FC<{ cluster: Cluster }> = ({ cluster }) => {
  const { ingressUrl, name } = cluster;
  return ingressUrl ? (
    <NameLink href={ingressUrl}>{name}</NameLink>
  ) : (
    <Name>{name}</Name>
  );
};

export const NotAvailable = styled.span`
  color: ${theme.colors.neutral40};
  font-style: italic;
  font-family: ${theme.fontFamilies.regular};
`;

const Status: FC<{
  status?: ClusterStatus;
  updatedAt?: string;
  connecting?: boolean;
}> = ({ status, updatedAt }) => {
  return status ? (
    <ReadyStatus
      status={getClusterStatus(status)}
      updatedAt={updatedAt}
      showConnectedStatus
    />
  ) : (
    <div>Unknown</div>
  );
};

export const statusBox = (cluster: Cluster, connecting: boolean) => {
  if (connecting && cluster.status === 'notConnected') {
    return (
      <Box
        lineHeight="24px"
        display="flex"
        flexDirection="column"
        alignItems="center"
        my={2}
      >
        <span>Waiting for connection from agent ...</span>
        <Loader />
      </Box>
    );
  }
  return (
    <Box lineHeight="24px" display="flex" my={2}>
      <Box color="text.secondary" mr={1}>
        Cluster status
      </Box>
      <Status
        connecting
        updatedAt={cluster.updatedAt}
        status={cluster.status}
      />
    </Box>
  );
};
export const Code = styled.div`
  display: flex;
  align-self: center;
  padding: 16px;
  background-color: ${theme.colors.white};
  font-family: ${theme.fontFamilies.monospace};
  border: 1px solid ${theme.colors.neutral20};
  border-radius: ${theme.borderRadius.soft};
  overflow: auto;
  font-size: ${theme.fontSizes.small};
`;

const TooltipStyle = styled.div`
  font-size: 14px;
`;

export const Tooltip: FC<TooltipProps & { disabled?: boolean }> = ({
  disabled,
  title,
  children,
  ...props
}) => {
  const styledTitle = <TooltipStyle>{title}</TooltipStyle>;
  return disabled ? (
    children
  ) : (
    <Mtooltip enterDelay={500} title={styledTitle} {...props}>
      {children}
    </Mtooltip>
  );
};

export const ColumnHeaderTooltip: FC<TooltipProps> = ({
  title,
  children,
  ...props
}) => (
  <Tooltip title={title} placement="top" {...props}>
    {children}
  </Tooltip>
);

const CommitMessage = styled.div`
  white-space: pre;
`;

const CommitContainer = styled.div`
  margin-bottom: 8px;
  line-height: 1.4em;
`;

const CommitHash = styled.span`
  font-family: ${theme.fontFamilies.monospace};
  font-size: 0.9em;
`;

const CommitAuthor = styled.span``;

interface CommitsTooltipProps {
  commits: GitCommitInfo[];
}
const CommitsTooltip: FC<CommitsTooltipProps> = ({ commits }) => (
  <div>
    {commits.map(commit => {
      const { author_name, author_date, message, sha } = commit;
      return (
        <CommitContainer key={sha}>
          <div>
            <CommitAuthor>{author_name}</CommitAuthor> commited{' '}
            <CommitHash>{sha.substring(0, 7)}</CommitHash>
          </div>
          <div>{author_date.Time}</div>
          <CommitMessage>{message}</CommitMessage>
        </CommitContainer>
      );
    })}
  </div>
);

const getGitCommitsUrl = (
  fluxInfo: FluxInfo,
  commits: GitCommitInfo[],
): string => {
  const parsedRepo = GitUrlParse(fluxInfo.repoUrl);
  if (commits.length === 1) {
    return `https://${parsedRepo.resource}/${parsedRepo.full_name}/commit/${commits[0].sha}`;
  }
  const sortedCommits = sortBy(commits, commit => commit.author_date.Time);
  const oldCommit = sortedCommits[0].sha;
  const recentCommit = sortedCommits[sortedCommits.length - 1].sha;

  return `https://${parsedRepo.resource}/${parsedRepo.full_name}/compare/${oldCommit}%5E...${recentCommit}`;
};

interface CommitsOverviewProps {
  fluxInfo?: FluxInfo;
  commits?: GitCommitInfo[];
}
export const CommitsOverview: FC<CommitsOverviewProps> = ({
  commits,
  fluxInfo,
}) => {
  if (!fluxInfo || !commits || commits.length === 0) {
    return <NotAvailable>Not available</NotAvailable>;
  }

  const data = commits.map(commit => ({
    ...commit,
    ts: new Date(commit.author_date.Time),
    // TODO: add 'success' | 'fail'
    status: '',
  }));

  const renderCommit = (element: JSX.Element, key: Key, data: any) => {
    const commits = data as GitCommitInfo[];
    return (
      <Tooltip
        enterDelay={250}
        key={key}
        title={<CommitsTooltip commits={commits} />}
      >
        <g onClick={() => window.open(getGitCommitsUrl(fluxInfo, commits))}>
          {element}
        </g>
      </Tooltip>
    );
  };

  return (
    <SparkTimeline
      renderCommit={renderCommit}
      showHeadLabel
      axisOnHover
      data={data}
    />
  );
};
