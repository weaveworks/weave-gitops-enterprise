import { createStyles, makeStyles } from '@material-ui/core';
import styled from 'styled-components';

export const usePipelineStyles = makeStyles(() =>
  createStyles({
    gridContainer: {
      margin: `12px 12px`,
      padding: '24px',
      borderRadius: '10px',
    },
    gridWrapper: {
      width: '100%',
      display: 'flex',
      flexWrap: 'nowrap',
      overflow: 'auto',
      paddingBottom: '8px',
      margin: `24 0 0 0`,
    },
    subtitleColor: {
      color: '#737373',
    },
    editButton: {
      paddingBottom: '12px',
    },
  }),
);
export const TargetWrapper = styled.div`
  font-size: ${props => props.theme.fontSizes.large};
  margin-bottom: ${props => props.theme.spacing.small};
  text-overflow: ellipsis;
  white-space: nowrap;
  overflow: hidden;
  width: calc(250px - ${props => props.theme.spacing.large});
`;
export const CardContainer = styled.div`
  background: ${props => props.theme.colors.white};
  padding: ${props => props.theme.spacing.small};
  margin-bottom: ${props => props.theme.spacing.xs};
  box-shadow: 0px 0px 1px rgba(26, 32, 36, 0.32);
  border-radius: 10px;
  font-weight: 600;
`;
export const Title = styled.div`
  font-size: ${props => props.theme.fontSizes.medium};
  color: ${props => props.theme.colors.black};
  font-weight: 400;
`;
export const ClusterName = styled.div`
  margin-bottom: ${props => props.theme.spacing.small};
  line-height: 24px;
  a > span {
    font-size: 20px;
  }
`;
export const TargetNamespace = styled.div`
  font-size: ${props => props.theme.fontSizes.medium};
`;
export const WorkloadWrapper = styled.div`
  position: relative;
  .version {
    margin-left: ${props => props.theme.spacing.xxs};
  }
`;
export const LastAppliedVersion = styled.span`
  color: ${props => props.theme.colors.neutral30};
  font-size: ${props => props.theme.fontSizes.medium};
  border: 1px solid ${props => props.theme.colors.neutral20};
  padding: 14px 6px;
  border-radius: 50%;
`;
