import { Box, CircularProgress } from '@material-ui/core';
import Alert from '@material-ui/lab/Alert';
import { Flex, Link, theme } from '@weaveworks/weave-gitops';
import { FC, useEffect, useState } from 'react';
import styled, { css } from 'styled-components';
import { ListError } from '../../cluster-services/cluster_services.pb';
import { useListVersion } from '../../hooks/versions';
import { NotificationData } from '../../types/custom';
import { Tooltip } from '../Shared';
import { AlertListErrors } from './AlertListErrors';
import Collapse from '@material-ui/core/Collapse';

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

const AlertWrapper = styled(Alert)`
  padding: ${base};
  margin: 0 ${base} ${base} ${base};
  border-radius: 10px;
  div[class*='MuiAlert-action'] {
    display: inline;
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
  notifications,
}) => {
  const { data, error } = useListVersion();
  const [open, setOpen] = useState<{ [index: number]: boolean }>({ 0: true });
  const [allNotifications, setAllNotifications] = useState<NotificationData[]>(
    [],
  );
  const entitlement = data?.entitlement;
  const versions = {
    capiServer: data?.data.version,
    ui: process.env.REACT_APP_VERSION || 'no version specified',
  };

  const handleOpen = (index: number) => {
    setOpen(prevState => ({
      ...prevState,
      [index]: !prevState[index],
    }));
  };

  useEffect(() => {
    let allNotif: NotificationData[] = [];
    if (entitlement) {
      allNotif = [
        ...allNotif,
        {
          message: { text: entitlement },
          severity: 'warning',
        },
      ];
    }
    if (error) {
      allNotif = [
        ...allNotif,
        {
          message: { text: error?.message },
          severity: 'error',
        },
      ];
    }
    if (notifications) {
      allNotif = [...allNotif, ...notifications];
    }

    allNotif.forEach((_, index) =>
      setOpen(prevState => ({
        ...prevState,
        [index]: true,
      })),
    );

    setAllNotifications(allNotif);
  }, [entitlement, error, notifications]);

  if (loading) {
    return (
      <Box marginTop={4}>
        <Flex wide center>
          <CircularProgress />
        </Flex>
      </Box>
    );
  }

  const notificationAlert = (n: NotificationData, index: number) => (
    <Box key={index}>
      <Collapse in={open[index]}>
        <AlertWrapper severity={n?.severity} onClose={() => handleOpen(index)}>
          {n?.message.text} {n?.message.component}
        </AlertWrapper>
      </Collapse>
    </Box>
  );

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
      {allNotifications
        ?.filter(n => n.display !== 'bottom')
        .map((n, index) => {
          return (
            (n?.message.text || n?.message.component) &&
            notificationAlert(n, index)
          );
        })}
      {type === 'WG' ? (
        <WGContent>{children}</WGContent>
      ) : (
        <Content backgroundColor={backgroundColor}>{children}</Content>
      )}
      {allNotifications
        ?.filter(n => n.display === 'bottom')
        .map((n, index) => {
          return (
            (n?.message.text || n?.message.component) &&
            notificationAlert(n, index)
          );
        })}
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
