import { Tab, Tabs, Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { theme } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { Dialog } from '@material-ui/core';

export const useWorkspaceStyle = makeStyles(() =>
  createStyles({
    navigateBtn: {
      marginBottom: theme.spacing.medium,
      marginRight: theme.spacing.none,
      textTransform: 'uppercase',
    },
    filterIcon: {
      color: theme.colors.primary10,
      marginRight: theme.spacing.small,
    },
  }),
);

export const WorkspacesTabs = styled(Tabs)`
  min-height: 32px !important;
  margin-top: ${({ theme }) => theme.spacing.medium};
  .link{
    color: ${({ theme }) => theme.colors.primary},
    fontWeight: 600,
    whiteSpace: 'pre-line',
  }
`;

export const WorkspaceTab = styled(Tab)(({ theme }) => ({
  '&.MuiTab-root': {
    fontSize: theme.fontSizes.small,
    fontWeight: 600,
    minHeight: '32px',
    minWidth: '133px',
    opacity: 1,
    paddingLeft: '0 !important',
    paddingRight: '0 !important',
    span: {
      color: theme.colors.neutral30,
    },
  },
  '&.Mui-selected': {
    fontWeight: 700,
    background: `${theme.colors.primary}1A`,
    span: {
      color: theme.colors.primary10,
    },
  },
  '&.Mui-focusVisible': {
    backgroundColor: '#d1eaff',
  },
}));

export const DialogWrapper = styled(Dialog)`
  .MuiDialog-paper {
    border-radius: 10px;
  }
  .MuiDialogTitle-root {
    background: ${({ theme }) => theme.colors.neutralGray};
    padding: ${({ theme }) => theme.spacing.medium};
    padding-bottom: ${({ theme }) => theme.spacing.small} ;
    p{
        font-weight: 600;
    }
    .MuiSvgIcon-root{
        color: ${({ theme }) => theme.colors.neutral30};
    }
    .info{
        color: ${({ theme }) => theme.colors.primary10} ;
        font-size: ${({ theme }) => theme.fontSizes.small};
        font-weight: 500;
    }
  }
  .MuiDialogContent-root{
    &.customBackgroundColor{
      background: ${({ theme }) => theme.colors.neutralGray} !important;
      padding:  ${({ theme }) => theme.spacing.none};
    }
    pre{
        background: ${({ theme }) => theme.colors.white}!important;
        padding-left:${({ theme }) => theme.spacing.none} !important;
        span{
        font-family: ${({ theme }) => theme.fontFamilies.monospace};
        font-size: ${({ theme }) => theme.fontSizes.small};
        text-align: left !important;
        padding-right: ${({ theme }) => theme.spacing.none} !important;
        min-width: 27px !important;
    }
  }
    }
  }
`;

export const RulesListWrapper = styled.ul`
  list-style: none;
  margin-top: ${({ theme }) => theme.spacing.none} !important;
  padding-left: ${({ theme }) => theme.spacing.none} !important;
  li {
    background: ${({ theme }) => theme.colors.white};
    margin-bottom: ${({ theme }) => theme.spacing.small};
    padding: ${({ theme }) => theme.spacing.small}
      ${({ theme }) => theme.spacing.medium};
    font-family: ${({ theme }) => theme.fontFamilies.monospace};
    font-size: ${({ theme }) => theme.fontSizes.small};
    label {
      margin-right: ${({ theme }) => theme.spacing.xs};
    }
  }
`;

export const ViewYamlBtn = styled.div`
  width: 100%;
  display: flex;
  justify-content: flex-end;
`;
