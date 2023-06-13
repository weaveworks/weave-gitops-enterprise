import { ReportProblem } from '@material-ui/icons';
import { Alert, Autocomplete } from '@material-ui/lab';
import { createStyles, makeStyles } from '@material-ui/styles';
import styled from 'styled-components';
import { TableWrapper } from '../Shared';

export const SectionTitle = styled.label`
display: block,
color: ${props => props.theme.colors.black},
font-size: ${props => props.theme.fontSizes.medium},
font-weight: 600,
margin-top: ${props => props.theme.spacing.large},
`;

export const TargetItemKind = styled.span`
  text-transform: capitalize;
  background: #eef0f4;
  padding: 4px 16px;
  color: ${props => props.theme.colors.black};
  margin-left: 12px;
  border-radius: 16px;
`;

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
    appliedTo: {
      marginTop: 16,
    },
    link: {
      color: '00b3ec',
      fontWeight: 600,
      whiteSpace: 'pre-line',
      textTransform: 'capitalize',
    },
    targetItemsList: {
      '& li': { marginTop: 8, display: 'flex', alignItems: 'center' },
      padding: 0,
      margin: 0,
    },
    targetItemKind: {
      background: '#eef0f4',
      padding: `${4} ${16}`,
      color: '#1a1a1a',
      marginLeft: 12,
      borderRadius: 16,
    },
    policyTitle: {
      '& span': {
        marginRight: 8,
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
      paddingLeft: 12,
      marginTop: 0,
      '& li': {
        width: '45%',
        '&.workspaces': {
          width: '33%',
          '& label': {
            marginBottom: '0 !important',
          },
        },
        '& .Mui-checked': {
          color: '009CCC',
        },
        '& svg': {
          marginRight: '5px',
        },
        '& label': {
          marginTop: `${8} !important`,
          marginBottom: `${16} !important`,
          fontSize: 20,
        },
      },
    },
    SelectPoliciesWithSearch: {
      '& div[class*="MuiOutlinedInput-root"]': {
        paddingTop: '0 !important',
        paddingBottom: '0 !important',
        paddingRight: `${12} !important`,
      },
      '& fieldset[class*="MuiOutlinedInput-root"]::hover': {
        borderColor: '#d8d8d8 !important',
      },
      '& div[class*="MuiFormControl-root"]': {
        paddingRight: 12,
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
    errorSection: {
      color: '#9F3119',
      display: 'Flex',
      alignItems: 'center',
      margin: 0,
      fontSize: 12,
      marginTop: 8,
      textAlign: 'left',
      fontWeight: 400,
      lineHeight: 1.66,
      '& svg': {
        marginRight: 4,
        width: '20px',
        height: '20px',
      },
    },
  }),
);

export const WarningIcon = styled(ReportProblem)`
  color: ${props => props.theme.colors.feedbackOriginal};
`;
export const WarningWrapper = styled(Alert)`
  background: ${props => props.theme.colors.feedbackLight} !important;
  margin-bottom: ${props => props.theme.spacing.small};
  height: 50px;
  border-radius: ${props => props.theme.spacing.xs} !important;
  color: ${props => props.theme.colors.black} !important;
  display: flex !important;
  align-items: center;
`;
export const PolicyConfigsTableWrapper = styled(TableWrapper)`
  table tbody tr td:first-child {
    width: 50px;
  }
`;

export const PolicyDetailsCardWrapper = styled.ul`
  padding-left: 0;
  list-style: 0;
  display: flex;
  flex-flow: wrap;
  li {
    width: 400px;
    padding: ${props => props.theme.spacing.small};
    .modified {
      color: #c2185b;
      display: block;
      margin-bottom: ${props => props.theme.spacing.xxs};
      font-size: ${props => props.theme.fontSizes.tiny};
      position: absolute;
      bottom: ${props => props.theme.spacing.xs};
    }
    .editPolicyCardHeader {
      justify-content: space-between;
      align-items: center;
      svg {
        color: ${props => props.theme.colors.alertMedium};
        cursor: pointer;
      }
    }
    .MuiCard-root {
      box-shadow: 0px 2px 8px 1px rgb(0 0 0 / 10%);
      border: 1px solid ${props => props.theme.colors.neutral20};
      height: 245px;
      border-radius: ${props => props.theme.spacing.xs} !important;
    }
    .cardLbl {
      color: ${props => props.theme.colors.neutral30};
      font-size: ${props => props.theme.fontSizes.small};
      display: block;
      font-weight: ${700};
      margin: ${props => props.theme.spacing.base} 0 0;
    }
    .parameterItem {
      font-size: ${props => props.theme.fontSizes.small};
      font-weight: 400;
      margin-top: ${props => props.theme.spacing.xs};
      position: relative;
      label {
        margin-bottom: ${props => props.theme.spacing.xs};
        display: block;
        font-size: ${props => props.theme.fontSizes.small};
        color: black;
      }

      label[class*='MuiFormControlLabel-root'] {
        height: 40px;
        display: flex;
        align-items: center;
        margin-bottom: ${props => props.theme.spacing.medium} !important;
        span[class*='PrivateSwitchBase-root'] {
          padding: 0 ${props => props.theme.spacing.xxs} 0
            ${props => props.theme.spacing.xs};
        }
        span {
          font-size: ${props => props.theme.fontSizes.small};
          svg {
            width: 20px;
            height: 20px;
          }
        }
      }

      .parameterItemValue {
        color: ${props => props.theme.colors.neutral30};
        label {
          padding-bottom: ${props => props.theme.spacing.xs};
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
    border: 0 !important;
  }
`;
