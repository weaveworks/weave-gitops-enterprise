import { createStyles, makeStyles } from '@material-ui/core';
import { theme } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

const { medium, xs, xxs, large } = theme.spacing;
const { small } = theme.fontSizes;
const { white, neutral10, neutral20, neutral30, black } = theme.colors;
export const usePipelineStyles = makeStyles(() =>
  createStyles({
    gridContainer: {
      backgroundColor: neutral10,
      margin: `0 ${small}`,
      padding: medium,
      borderRadius: '10px',
    },
    gridWrapper: {
      width: '100%',
      display: 'flex',
      flexWrap: 'nowrap',
      overflow: 'auto',
      paddingBottom: '8px',
      margin: `${medium} 0 0 0`,
    },
    title: {
      fontSize: `calc(${small} + ${small})`,
      fontWeight: 600,
      textTransform: 'capitalize',
    },
    subtitle: {
      fontSize: small,
      fontWeight: 400,
      marginTop: xs,
    },
    mbSmall: {
      marginBottom: small,
    },
    subtitleColor: {
      color: neutral30,
    },
    editButton: {
      paddingBottom: theme.spacing.small,
    },
  }),
);
export const TargetWrapper = styled.div`
  font-size: ${theme.fontSizes.large};
  margin-bottom: ${small};
  text-overflow: ellipsis;
  white-space: nowrap;
  overflow: hidden;
  width: calc(250px - ${large});
`;
export const CardContainer = styled.div`
  background: ${white};
  padding: ${small};
  margin-bottom: ${xs};
  box-shadow: 0px 0px 1px rgba(26, 32, 36, 0.32);
  border-radius: 10px;
  font-weight: 600;
`;
export const Title = styled.div`
  font-size: ${theme.fontSizes.medium};
  color: ${black};
  font-weight: 400;
`;
export const ClusterName = styled.div`
  margin-bottom: ${small};
  line-height: 24px;
  a > span {
    font-size: 20px;
  }
`;
export const TargetNamespace = styled.div`
  font-size: ${theme.fontSizes.medium};
`;
export const WorkloadWrapper = styled.div`
  position: relative;
  .version {
    margin-left: ${xxs};
  }
`;
export const LastAppliedVersion = styled.span`
  color: ${neutral30};
  font-size: ${theme.fontSizes.medium};
  border: 1px solid ${neutral20};
  padding: 14px 6px;
  border-radius: 50%;
`;
