import { Box, CircularProgress } from '@material-ui/core';
import Alert from '@material-ui/lab/Alert';
import { createStyles, makeStyles } from '@material-ui/styles';
import { Flex, theme } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import styled, { css } from 'styled-components';
import { ListError } from '../../cluster-services/cluster_services.pb';
import { useListVersion } from '../../hooks/versions';
import { Tooltip } from '../Shared';
import useNotifications from './../../contexts/Notifications';
import { AlertListErrors } from './AlertListErrors';

const xs = theme.spacing.xs;
const small = theme.spacing.small;
const medium = theme.spacing.medium;
const large = theme.spacing.large;

export const Title = styled.h2`
  margin-top: 0px;
`;

export const pageDimensionsCss = css`
  width: 100%;
`;

export const PageWrapper = styled.div`
  ${pageDimensionsCss}
  margin: 0 auto;
`;

export const contentCss = css`
  margin: ${medium} ${small} 0 ${small};
  padding: ${large} ${medium} ${medium} ${large};
  background-color: ${theme.colors.white};
  border-radius: ${xs};
`;

export const Content = styled.div<{ backgroundColor?: string }>`
  ${contentCss};
  background-color: ${props => props.backgroundColor};
`;

export const WGContent = styled.div`
  margin: ${medium} ${small} 0 ${small};
  background-color: ${theme.colors.white};
  border-radius: ${xs};
  > div > div {
    border-radius: ${xs};
    max-width: none;
  }
`;

const HelpLinkWrapper = styled.div`
  padding: ${small} ${medium};
  margin: ${small};
  background-color: ${theme.colors.white};
  color: ${({ theme }) => theme.colors.neutral40};
  border-radius: ${xs};
  display: flex;
  justify-content: space-between;
  a {
    color: ${({ theme }) => theme.colors.primary};
  }
`;

const useStyles = makeStyles(() =>
  createStyles({
    alertWrapper: {
      marginTop: theme.spacing.medium,
      marginRight: theme.spacing.small,
      marginBottom: 0,
      marginLeft: theme.spacing.small,
      paddingRight: theme.spacing.medium,
      paddingLeft: theme.spacing.medium,
      borderRadius: theme.spacing.xs,
    },
    warning: {
      backgroundColor: theme.colors.feedbackLight,
    },
  }),
);

interface Props {
  type?: string;
  backgroundColor?: string;
  errors?: ListError[];
  loading?: boolean;
}

export const ContentWrapper: FC<Props> = ({
  children,
  type,
  backgroundColor,
  errors,
  loading,
}) => {
  const classes = useStyles();
  const { setNotifications } = useNotifications();
  const { data, error } = useListVersion();
  const entitlement = data?.entitlement;
  const versions = {
    capiServer: data?.data.version,
    ui: process.env.REACT_APP_VERSION || 'no version specified',
  };

  if (loading) {
    return (
      <Box marginTop={4}>
        <Flex wide center>
          <CircularProgress />
        </Flex>
      </Box>
    );
  }

  if (error) {
    setNotifications([{ message: { text: error.message }, variant: 'danger' }]);
  }

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        width: '100%',
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
      <AlertListErrors errors={errors} />
      {type === 'WG' ? (
        <WGContent>{children}</WGContent>
      ) : (
        <Content backgroundColor={backgroundColor}>{children}</Content>
      )}
      <HelpLinkWrapper>
        <div>
          Need help? Contact us at&nbsp;
          <a href="mailto:support@weave.works">support@weave.works</a>
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
