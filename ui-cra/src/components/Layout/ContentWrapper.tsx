import React, { FC } from 'react';
import styled, { css } from 'styled-components';
import { theme } from '@weaveworks/weave-gitops';
import { Tooltip } from '../Shared';
import { ListError } from '../../cluster-services/cluster_services.pb';
import Alert from '@material-ui/lab/Alert';
import AlertTitle from '@material-ui/lab/AlertTitle';
import { createStyles, makeStyles } from '@material-ui/styles';
import { ListItem } from '@material-ui/core';
import { useListVersion } from '../../hooks/versions';
import useNotifications from './../../contexts/Notifications';

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
  margin: 0 ${small};
  background-color: ${theme.colors.white};
  color: ${({ theme }) => theme.colors.neutral40};
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

export const ContentWrapper: FC<{
  type?: string;
  backgroundColor?: string;
  errors?: ListError[];
}> = ({ children, type, backgroundColor, errors }) => {
  const classes = useStyles();
  const { setNotifications } = useNotifications();
  const { data, error } = useListVersion();
  const entitlement = data?.entitlement;
  const versions = {
    capiServer: data?.data.version,
    ui: process.env.REACT_APP_VERSION || 'no version specified',
  };

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
      {!!(errors && errors.length) && (
        <Alert className={classes.alertWrapper} severity="error">
          <AlertTitle>
            There was a problem retrieving results from some clusters:
          </AlertTitle>
          {errors?.map((item: ListError) => (
            <ListItem key={item.clusterName}>
              - Cluster {item.clusterName} {item.message}
            </ListItem>
          ))}
        </Alert>
      )}
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
