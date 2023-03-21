import { ReportProblem } from '@material-ui/icons';
import { Alert } from '@material-ui/lab';
import { createStyles, makeStyles } from '@material-ui/styles';
import { theme } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { TableWrapper } from '../Shared';

const {
  defaultOriginal,
  black,
  primary,
  neutral30,
  neutral20,
  feedbackLight,
  backGrey,
} = theme.colors;
const { medium: mediumFont } = theme.fontSizes;
const { none, xxs, xs, small, base, large } = theme.spacing;

export const usePolicyConfigStyle = makeStyles(() =>
  createStyles({
    centered: {
      textAlign: 'center',
      width: '100px',
    },
    capitlize: {
      textTransform: 'capitalize',
    },
    upperCase: {
      textTransform: 'uppercase',
    },
    sectionTitle: {
      color: black,
      fontSize: mediumFont,
      fontWeight: 600,
      marginTop: large,
      display: 'block',
    },
    appliedTo: {
      marginTop: base,
    },
    link: {
      color: primary,
      fontWeight: 600,
      whiteSpace: 'pre-line',
      textTransform: 'capitalize',
    },
    targetItemsList: {
      '& li': { marginTop: xs, display: 'flex', alignItems: 'center' },
      padding: none,
      margin: none,
    },
    targetItemKind: {
      background: backGrey,
      padding: `${xxs} ${base}`,
      color: black,
      marginLeft: small,
      borderRadius: base,
    },
    policyTitle: {
      '& span': {
        marginRight: xs,
      },
      display: 'flex',
      alignItems: 'flex-start',
      whiteSpace: 'pre-line',
      textTransform: 'capitalize',
    },
  }),
);

export const WarningIcon = styled(ReportProblem)`
  color: ${defaultOriginal};
`;
export const WarningWrapper = styled(Alert)`
  background: ${feedbackLight} !important;
  margin-bottom: ${small};
  height: 50px;
  border-radius: ${xs} !important;
  color: ${black} !important;
  display: flex !important;
  align-items: center;
`;
export const PolicyConfigsTableWrapper = styled(TableWrapper)`
  table tbody tr td:first-child {
    width: 50px;
  }
`;

export const PolicyDetailsCardWrapper = styled.ul`
  padding-left: ${none};
  list-style: none;
  display: flex;
  flex-flow: wrap;
  li {
    width: 33%;
    padding: ${small};
    .MuiCard-root {
      box-shadow: 0px 2px 8px 1px rgb(0 0 0 / 10%);
      border: 1px solid ${neutral20};
      height: 245px;
      border-radius: ${xs} !important;
    }
    .cardLbl {
      color: ${neutral30};
      font-size: ${small};
      display: block;
      font-weight: ${700};
      margin: ${base} ${none} ${none};
    }
    .parameterItem {
      font-size: ${small};
      font-weight: 400;
      margin-top: ${xs};
      label {
        margin-bottom: ${xs};
        display: block;
      }
      .parameterItemValue {
        color: ${neutral30};
      }
    }
  }
`;
