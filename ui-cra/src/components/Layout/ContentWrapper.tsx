import { Box, CircularProgress } from '@material-ui/core';
import { Flex } from '@weaveworks/weave-gitops';
import { FC, useEffect } from 'react';
import styled, { css } from 'styled-components';
import { ListError } from '../../cluster-services/cluster_services.pb';

import useNotifications, {
  NotificationData,
} from './../../contexts/Notifications';
import { AlertListErrors } from './AlertListErrors';
import Notifications from './Notifications';

import { useVersionContext } from '../../contexts/ListConfig';
import {
  WarningIcon,
  WarningWrapper,
} from '../PolicyConfigs/PolicyConfigStyles';
import MemoizedHelpLinkWrapper from './HelpLinkWrapper';

const ENTITLEMENT_ERROR =
  'No entitlement was found for Weave GitOps Enterprise. Please contact sales@weave.works.';

const ENTITLEMENT_WARN =
  'Your entitlement for Weave GitOps Enterprise has expired, please contact sales@weave.works.';

export const Title = styled.h2`
  margin-top: 0px;
`;

export const PageWrapper = styled.div`
  width: 100%;
  height: 100%;
  margin: 0 auto;
`;

export const contentCss = css`
  padding: ${props => props.theme.spacing.medium};
  background-color: ${props => props.theme.colors.white};
  border-radius: ${props => props.theme.spacing.xs}
    ${props => props.theme.spacing.xs} 0 0;
  height: 100%;
`;

export const Content = styled.div<{ backgroundColor?: string }>`
  ${contentCss};
  background-color: ${props => props.backgroundColor};
`;

interface Props {
  type?: string;
  backgroundColor?: string;
  errors?: ListError[];
  loading?: boolean;
  notifications?: NotificationData[];
  warningMsg?: string;
}

export const ContentWrapper: FC<Props> = ({
  children,
  backgroundColor,
  errors,
  loading,
  warningMsg,
}) => {
  const versionResponse = useVersionContext();
  const { notifications, setNotifications } = useNotifications();

  useEffect(() => {
    if (versionResponse?.entitlement === ENTITLEMENT_WARN) {
      setNotifications([
        {
          message: {
            text: versionResponse.entitlement,
          },
          severity: 'warning',
        } as NotificationData,
      ]);
    }
  }, [versionResponse?.entitlement, setNotifications]);

  const topNotifications = notifications.filter(
    n => n.display !== 'bottom' && n.message.text !== ENTITLEMENT_ERROR,
  );
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
      id="content-wrapper"
      style={{
        display: 'flex',
        flexDirection: 'column',
        width: '100%',
        maxHeight: 'calc(100vh - 60px)',
        overflowWrap: 'normal',
        overflowX: 'scroll',
        padding: '0 24px',
        margin: '0 auto',
      }}
    >
      {errors && (
        <AlertListErrors
          errors={errors.filter(error => error.message !== ENTITLEMENT_ERROR)}
        />
      )}
      {!!warningMsg && (
        <WarningWrapper
          severity="warning"
          iconMapping={{
            warning: <WarningIcon />,
          }}
        >
          <span>{warningMsg}</span>
        </WarningWrapper>
      )}
      <Notifications notifications={topNotifications} />

      <Content backgroundColor={backgroundColor}>{children}</Content>

      {!!bottomNotifications.length && (
        <div style={{ paddingTop: '16px' }}>
          <Notifications notifications={bottomNotifications} />
        </div>
      )}
      <MemoizedHelpLinkWrapper />
    </div>
  );
};
