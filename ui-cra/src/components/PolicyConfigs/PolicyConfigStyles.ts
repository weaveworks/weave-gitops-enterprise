import { ReportProblem } from '@material-ui/icons';
import { Alert, Autocomplete } from '@material-ui/lab';
import { createStyles, makeStyles } from '@material-ui/styles';
import { theme } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { TableWrapper } from '../Shared';

const {
  defaultOriginal,
  black,
  primary,
  primary10,
  neutral30,
  neutral20,
  feedbackLight,
  backGrey,
  alertMedium,
} = theme.colors;
const {
  medium: mediumFont,
  small: smallFont,
  large: largeFontSize,
  tiny: tinyFont,
} = theme.fontSizes;
const { none, xxs, xs, small, base, large, medium } = theme.spacing;
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
    checkList: {
      display: 'flex',
      listStyle: 'none',
      flexFlow: 'wrap',
      paddingLeft: small,
      '& li': {
        width: '45%',
        '& .Mui-checked': {
          color: primary10,
        },
        '& svg': {
          marginRight: '5px',
        },
        '& label': {
          marginTop: `${xs} !important`,
          marginBottom: `${base} !important`,
          fontSize: largeFontSize,
        },
      },
    },
    SelectPoliciesWithSearch: {
      '& div[class*="MuiOutlinedInput-root"]': {
        paddingTop: '0 !important',
        paddingBottom: '0 !important',
        paddingRight: `${small} !important`,
      },
      '& fieldset[class*="MuiOutlinedInput-root"]::hover': {
        borderColor: '#d8d8d8 !important',
      },
      '& div[class*="MuiFormControl-root"]': {
        paddingRight: small,
      },

      '& div[class*="MuiChip-root"]': {
        height: '26px',
      },
      '& input': {
        border: 'none !important',
      },
      '& svg': {
        color: '#0000008a !important',
      },
    },
    fieldNote: {
      textTransform: 'uppercase',
      marginBottom: small,
      display: 'block',
      color: neutral30,
      fontSize: smallFont,
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
    width: 400px;
    padding: ${small};
    .modified {
      color: #c2185b;
      display: block;
      margin-bottom: ${xxs};
      font-size: ${tinyFont};
      position: absolute;
      bottom: ${xxs};
    }
    .editPolicyCardHeader {
      justify-content: space-between;
      align-items: center;
      svg {
        color: ${alertMedium};
        cursor: pointer;
      }
    }
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
      position: relative;
      label {
        margin-bottom: ${xs};
        display: block;
        font-size: ${small};
        color: black;
      }

      label[class*='MuiFormControlLabel-root'] {
        height: 40px;
        display: flex;
        align-items: center;
        margin-bottom: ${medium} !important;
        span[class*='PrivateSwitchBase-root'] {
          padding: 0 ${xxs} 0 ${xs};
        }
        span {
          font-size: ${small};
          svg {
            width: 20px;
            height: 20px;
          }
        }
      }

      .parameterItemValue {
        color: ${neutral30};
        label {
          padding-bottom: ${xs};
        }
      }
    }
  }
`;

export const SelectPoliciesWithSearch = styled(Autocomplete)`
  div[class*='MuiOutlinedInput-root'] {
    padding-top: 0 !important;
    padding-bottom: 0 !important;
  }
  input {
    border: none !important;
  }
`;
