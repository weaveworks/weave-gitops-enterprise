import React, { FC } from 'react';
import { Box, CircularProgress } from '@material-ui/core';
import { Flex, Link, theme } from '@weaveworks/weave-gitops';
import styled, { css } from 'styled-components';
import { ListError } from '../../cluster-services/cluster_services.pb';
import { useListVersion } from '../../hooks/versions';
import { Tooltip } from '../Shared';
import { AlertListErrors } from './AlertListErrors';
import useNotifications, {
  NotificationData,
} from './../../contexts/Notifications';
import Notifications from './Notifications';

const { xxs, xs, small, medium, base } = theme.spacing;
const { white } = theme.colors;

export const Title = styled.h2`
  margin-top: 0px;
`;

export const PageWrapper = styled.div`
  width: 100%;
  margin: 0 auto;
`;

export const contentCss = css`
  margin: 0 ${base};
  padding: ${medium};
  background-color: ${white};
  border-radius: ${xs} ${xs} 0 0;
`;

export const Content = styled.div<{ backgroundColor?: string }>`
  ${contentCss};
  background-color: ${props => props.backgroundColor};
`;

export const WGContent = styled.div`
  margin: ${medium} ${small} 0 ${small};
  background-color: ${white};
  border-radius: ${xs} ${xs} 0 0;
  > div > div {
    border-radius: ${xs};
    max-width: none;
  }
`;

const HelpLinkWrapper = styled.div`
  padding: calc(${medium} - ${xxs}) ${medium};
  margin: 0 ${base};
  background-color: rgba(255, 255, 255, 0.7);
  color: ${({ theme }) => theme.colors.neutral30};
  border-radius: 0 0 ${xs} ${xs};
  display: flex;
  justify-content: space-between;
  a {
    color: ${({ theme }) => theme.colors.primary};
  }
`;

interface Props {
  type?: string;
  backgroundColor?: string;
  errors?: ListError[];
  loading?: boolean;
  notifications?: NotificationData[];
}

export const ContentWrapper: FC<Props> = ({
  children,
  type,
  backgroundColor,
  errors,
  loading,
}) => {
  const { data, error } = useListVersion();
  const { notifications } = useNotifications();
  const entitlement = data?.entitlement;
  const versions = {
    capiServer: data?.data.version,
    ui: process.env.REACT_APP_VERSION || 'no version specified',
  };

  const topNotifications = notifications.filter(n => n.display !== 'bottom');
  const bottomNotifications = notifications.filter(n => n.display === 'bottom');

  if (loading) {
    return (
      <Box marginTop={4}>
        <Flex wide center>
          <CircularProgress />
        </Flex>
      </Box>
    );
  }

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        width: '100%',
        height: 'calc(100vh - 80px)',
        overflowWrap: 'normal',
        overflowX: 'scroll',
      }}
    >
      {errors && <AlertListErrors errors={errors} />}
      <Notifications
        notifications={[
          ...topNotifications,
          {
            message: { text: entitlement },
            severity: 'warning',
          } as NotificationData,
          { message: { text: error?.message }, severity: 'error' },
        ]}
      />
      {type === 'WG' ? (
        <WGContent>{children}</WGContent>
      ) : (
        <Content backgroundColor={backgroundColor}>{children}</Content>
      )}
      <Notifications notifications={bottomNotifications} />
      <HelpLinkWrapper>
        <div>
          Need help? Raise a&nbsp;
          <Link newTab href="https://weavesupport.zendesk.com/">
            support ticket
          </Link>
        </div>
        <Tooltip
          title={`Server Version ${versions?.capiServer}`}
          placement="top"
        >
          <div>Weave GitOps Enterprise {process.env.REACT_APP_VERSION}</div>
        </Tooltip>
      </HelpLinkWrapper>
    </div>
  );
};
