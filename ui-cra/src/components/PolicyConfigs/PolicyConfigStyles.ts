import { ReportProblem } from '@material-ui/icons';
import { createStyles, makeStyles } from '@material-ui/styles';
import { theme } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

const { defaultOriginal } = theme.colors;


export const usePolicyConfigStyle = makeStyles(() =>
  createStyles({
   centered:{
    textAlign: 'center',
    width: '100px'
   }
  }),
);

export const WarningIcon = styled(ReportProblem)`
  color: ${defaultOriginal};
`;