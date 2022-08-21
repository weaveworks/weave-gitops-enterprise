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

const { xs, small, medium, base } = theme.spacing;
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
  padding: ${small} ${medium};
  margin: 0 ${base};
  background-color: rgba(255, 255, 255, 0.85);
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
      marginTop: medium,
      marginRight: small,
      marginBottom: 0,
      marginLeft: small,
      paddingRight: medium,
      paddingLeft: medium,
      borderRadius: xs,
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
      <AlertListErrors errors={errors} />
      {type === 'WG' ? (
        <WGContent>{children}</WGContent>
      ) : (
        <Content backgroundColor={backgroundColor}>{children}</Content>
      )}
      <HelpLinkWrapper>
        <div>
          Need help? Raise a&nbsp;
          <a href="https://weavesupport.zendesk.com/">support ticket</a>
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
