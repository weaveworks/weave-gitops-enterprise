import { ReportProblem } from '@material-ui/icons';
import { Alert, Autocomplete } from '@material-ui/lab';
import { createStyles, makeStyles } from '@material-ui/styles';
import styled from 'styled-components';
import { TableWrapper } from '../Shared';
import { RemoveCircleOutline } from '@material-ui/icons';

export const SectionTitle = styled.label`
  display: block;
  color: ${props => props.theme.colors.black};
  font-size: ${props => props.theme.fontSizes.medium};
  font-weight: 600;
  margin-top: ${props => props.theme.spacing.large};
`;

export const TargetItemKind = styled.span`
  text-transform: capitalize;
  background: ${props => props.theme.colors.neutralGray};
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
  list-style: none;
  display: flex;
  flex-flow: wrap;
  li {
    width: 24%;
    padding: ${props => props.theme.spacing.base}
      ${props => props.theme.spacing.small};
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
      background: ${props => props.theme.colors.neutralGray};

      box-shadow: 0px 2px 8px 1px rgb(0 0 0 / 10%);
      border: 1px solid ${props => props.theme.colors.neutral20};
      min-height: 245px;
      height: 100%;
      border-radius: ${props => props.theme.spacing.xs} !important;
    }
    .cardLbl {
      color: ${props => props.theme.colors.black};
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
        color: ${props => props.theme.colors.black};
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
export const CheckList = styled.ul`
  display:flex;
  min-width: 100vh;  
  list-style: none;
  flex-flow: wrap;
  padding-left:   ${props => props.theme.spacing.small};
  margin-top:  ${props => props.theme.spacing.none};
   li{
    width: 49%;
    &.workspaces {
      width: 33%;
    },
    svg {
      margin-right: ${props => props.theme.spacing.xxs};
    },
},`;
export const ErrorSection = styled.div`
color: ${props => props.theme.colors.alertDark};
display: Flex;
align-items: center;
margin: ${props => props.theme.spacing.none};
font-size: ${props => props.theme.fontSizes.small};
margin-top:  ${props => props.theme.spacing.xs};
text-align: left;
font-weight: 400;
line-height: 1.66;
svg {
  margin-right: ${props => props.theme.spacing.xxs};
  width: 20px;
  height: 20px;
},
`;

export const RemoveIcon = styled(RemoveCircleOutline)`
  color: ${props => props.theme.colors.alertMedium};
`;
