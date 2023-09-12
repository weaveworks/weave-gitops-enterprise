import { Dialog } from '@material-ui/core';
import { SubRouterTabs } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

export const CustomSubRouterTabs = styled(SubRouterTabs)(props => ({
  '.MuiTabs-root': {
    marginTop: props.theme.spacing.medium,
    width: '100%',
  },
}));

export const DialogWrapper = styled(Dialog)`
  .MuiDialog-paper {
    border-radius: 10px;
  }
  .MuiDialogTitle-root {
    background: ${props => props.theme.colors.neutralGray};
    padding: ${props => props.theme.spacing.medium};
    padding-bottom: ${props => props.theme.spacing.small} ;
    p{
        font-weight: 600;
    }
    .MuiSvgIcon-root{
        color: ${props => props.theme.colors.neutral30};
    }
    .info{
        color: ${props => props.theme.colors.primary10} ;
        font-size: ${props => props.theme.fontSizes.small};
        font-weight: 500;
    }
  }
  .MuiDialogContent-root{
    &.customBackgroundColor{
      background: ${props => props.theme.colors.neutralGray} !important;
      padding: 0;
    }
    pre{
        background: ${props => props.theme.colors.white} !important;
        padding-left:0 !important;
        span{
        font-family: ${props => props.theme.fontFamilies.monospace};
        font-size: ${props => props.theme.fontSizes.small};
        text-align: left !important;
        padding-right: 0 !important;
        min-width: 27px !important;
    }
  }
    }
  }
`;

export const RulesListWrapper = styled.ul`
  list-style: none;
  margin-top: 0 !important;
  padding-left: 0 !important;
  li {
    background: ${props => props.theme.colors.white};
    margin-bottom: ${props => props.theme.spacing.small};
    padding: ${props =>
      props.theme.spacing.small + ' ' + props.theme.spacing.medium};
    font-family: ${props => props.theme.fontFamilies.monospace};
    font-size: ${props => props.theme.fontSizes.small};
    label {
      margin-right: ${props => props.theme.spacing.xs};
    }
  }
`;
