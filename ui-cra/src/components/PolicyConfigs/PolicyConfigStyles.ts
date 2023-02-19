import { ReportProblem } from '@material-ui/icons';
import { createStyles, makeStyles } from '@material-ui/styles';
import { theme } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { TableWrapper } from '../Shared';

const { defaultOriginal } = theme.colors;

export const usePolicyConfigStyle = makeStyles(() =>
  createStyles({
    centered: {
      textAlign: 'center',
      width: '100px',
    },
    fixedCell: {},
  }),
);

export const WarningIcon = styled(ReportProblem)`
  color: ${defaultOriginal};
`;

export const PolicyConfigsTableWrapper = styled(TableWrapper)`
  table tbody tr td:first-child {
    width: 50px;
  }
`;
