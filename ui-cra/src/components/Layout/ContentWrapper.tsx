import { Box, CircularProgress } from '@material-ui/core';
import Alert from '@material-ui/lab/Alert';
import { createStyles, makeStyles } from '@material-ui/styles';
import { Flex, Link, theme } from '@weaveworks/weave-gitops';
import { FC, useEffect } from 'react';
import styled, { css } from 'styled-components';
import { ListError } from '../../cluster-services/cluster_services.pb';
import { useListVersion } from '../../hooks/versions';
import { Tooltip } from '../Shared';
import useNotifications from './../../contexts/Notifications';
import { AlertListErrors } from './AlertListErrors';

const { xxs, xs, small, medium, base } = theme.spacing;
const { feedbackLight, white } = theme.colors;

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

const useStyles = makeStyles(() =>
  createStyles({
    alertWrapper: {
      padding: base,
      margin: `0 ${base} ${base} ${base}`,
      borderRadius: '10px',
    },
    warning: {
      backgroundColor: feedbackLight,
    },
  }),
);

interface Props {
  type?: string;
  backgroundColor?: string;
  errors?: ListError[];
  loading?: boolean;
  errorMessage?: string;
}

export const ContentWrapper: FC<Props> = ({
  children,
  type,
  backgroundColor,
  errors,
  loading,
  errorMessage,
}) => {
  const classes = useStyles();
  const { setNotifications } = useNotifications();
  const { data, error } = useListVersion();
  const entitlement = data?.entitlement;
  const versions = {
    capiServer: data?.data.version,
    ui: process.env.REACT_APP_VERSION || 'no version specified',
  };

  useEffect(() => {
    if (error) {
      setNotifications([
        { message: { text: error?.message }, variant: 'danger' },
      ]);
    }
  }, [error, setNotifications]);

  if (loading) {
    return (
      <Box marginTop={4}>
        <Flex wide center>
          <CircularProgress />
        </Flex>
      </Box>
    );
  }
  if (errorMessage) {
    return (
      <Alert severity="error" className={classes.alertWrapper}>
        {errorMessage}
      </Alert>
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
      {entitlement && (
        <Alert
          className={`${classes.alertWrapper} ${classes.warning}`}
          severity="warning"
        >
          {entitlement}
        </Alert>
      )}
      {errors && <AlertListErrors errors={errors} />}
      {type === 'WG' ? (
        <WGContent>{children}</WGContent>
      ) : (
        <Content backgroundColor={backgroundColor}>{children}</Content>
      )}
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
