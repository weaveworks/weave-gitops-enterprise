import { RemoveCircleOutline, ReportProblem } from '@material-ui/icons';
import { Alert, Autocomplete } from '@material-ui/lab';
import { createStyles, makeStyles } from '@material-ui/styles';
import { Flex, Text } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { TableWrapper } from '../Shared';

export const TargetItemKind = styled(Text)`
  background: ${props => props.theme.colors.neutralGray};
  padding: ${props => props.theme.spacing.xxs}
    ${props => props.theme.spacing.base};
  color: ${props => props.theme.colors.black};
  margin-left: ${props => props.theme.spacing.small};
  border-radius: ${props => props.theme.spacing.base};
`;
export const TotalPolicies = styled(Flex)`
  width: 100px;
`;
export const usePolicyConfigStyle = makeStyles(() =>
  createStyles({
    capitlize: {
      textTransform: 'capitalize',
    },
    upperCase: {
      textTransform: 'uppercase',
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
  padding-right: 0 !important;
  padding-left: 0 !important;
  .MuiAlert-icon {
    margin-left: ${props => props.theme.spacing.base} !important;
  }
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
  &.policyDetails {
    width: 80%;
  }
  li {
    width: 30%;
    padding: ${props => props.theme.spacing.base}
      ${props => props.theme.spacing.small};
    .modified {
      color: #c2185b;
      display: block;
      margin-bottom: ${props => props.theme.spacing.none} !important;
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
    .MuiInputBase-root {
      border: 1px solid ${props => props.theme.colors.neutral20};
      margin-right: ${props => props.theme.spacing.none};
    }
    .MuiFormControl-root {
      width: 100%;
      .MuiFormLabel-root {
        font-size: ${props => props.theme.fontSizes.small};
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
      margin-top: ${props => props.theme.spacing.base};
    }
    .parameterItem {
      font-size: ${props => props.theme.fontSizes.small};
      font-weight: 400;
      margin-top: ${props => props.theme.spacing.xs};
      position: relative;
      span {
        display: block;
        margin-bottom: ${props => props.theme.spacing.xs};
      }

      label[class*='MuiFormControlLabel-root'] {
        margin-top: ${props => props.theme.spacing.xs};
        margin-bottom: ${props => props.theme.spacing.xs};

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
  cursor: pointer;
`;
