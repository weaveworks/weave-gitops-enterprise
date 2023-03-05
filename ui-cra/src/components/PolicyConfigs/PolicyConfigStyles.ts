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
  feedbackLight,
  neutralGray,
} = theme.colors;
const { medium: mediumFont } = theme.fontSizes;
const { none, xs, small , base} = theme.spacing;
console.log(theme.borderRadius);
export const usePolicyConfigStyle = makeStyles(() =>
  createStyles({
    centered: {
      textAlign: 'center',
      width: '100px',
    },
    capitlize: {
      textTransform: 'capitalize',
    },
    sectionTitle: {
      color: black,
      fontSize: mediumFont,
      fontWeight: 600,
    },
    link: {
      color: primary,
      fontWeight: 600,
      whiteSpace: 'pre-line',
    },
    targetItemsList: {
      '& li': { marginTop: small },
      listStyle: 'none',
      padding: none,
    },
    targetItemKind: {
      background: neutralGray,
      padding: small,
      color: black,
      marginLeft: small,
      borderRadius: base
    },
  }),
);

export const WarningIcon = styled(ReportProblem)`
  color: ${defaultOriginal};
`;
export const WarningWrapper = styled(Alert)`
  background: ${feedbackLight} !important;
  margin: ${none} ${small} ${small};
  height: 50px;
  border-radius: ${xs} !important;
  font-weight: 600 !important;
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
    width: 25%;
    padding: ${small};
    .MuiCard-root {
      box-shadow: 0px 2px 8px 1px rgb(0 0 0 / 10%);
      height: 245px;
      border-radius: ${xs} !important;
    }
    .cardLbl {
      color: ${neutral30};
      font-size: ${small};
      display: block;
      font-weight: ${700};
      margin: ${small} ${none};
    }
    .parameterItem {
      font-size: ${small};
      font-weight: 400;
      margin-bottom: ${xs};
      label {
        margin-bottom: ${xs};
        display: block;
      }
      .parameterItemValue {
        color: ${neutral30};
        text-transform: capitalize;

      }
    }
  }
`;
