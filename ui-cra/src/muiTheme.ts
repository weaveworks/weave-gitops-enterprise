import { createTheme } from '@material-ui/core/styles';
import {
  muiTheme as coreMuiTheme,
  theme as weaveTheme,
} from '@weaveworks/weave-gitops';

const defaultTheme = createTheme();

export const muiTheme = createTheme({
  ...coreMuiTheme,
  overrides: {
    MuiButton: {
      root: {
        textTransform: 'none',
        minWidth: 52,
      },
    },
    MuiDialog: {
      root: {
        padding: 0,
      },
      paper: {
        padding: 0,
        backgroundColor: weaveTheme.colors.white,
      },
    },
    MuiDialogActions: {
      root: {
        margin: 0,
        padding: defaultTheme.spacing(0, 2, 2, 2),
        justifyContent: 'flex-end',
      },
    },
    MuiDialogTitle: {
      root: {
        padding: defaultTheme.spacing(2, 2, 0, 2),
      },
    },
    MuiDialogContent: {
      root: {
        padding: defaultTheme.spacing(1, 2, 2, 2),
      },
    },
    MuiInputBase: {
      root: {
        flexGrow: 1,
      },
      input: {
        border: `1px solid ${weaveTheme.colors.neutral20}`,
        borderRadius: 2,
        position: 'relative',
        backgroundColor: defaultTheme.palette.common.white,
        fontSize: 16,
        width: '100%',
        padding: '8px 12px',
        '&:focus': {
          borderColor: weaveTheme.colors.primaryDark,
          borderRadius: 2,
        },
      },
    },
    MuiSelect: {
      select: {
        width: '100%',
      },
    },
    MuiTableCell: {
      head: {
        color: weaveTheme.colors.neutral30,
      },
      root: {
        borderBottom: 'none',
      },
    },
    MuiTablePagination: {
      select: {
        fontSize: 14,
      },
      toolbar: {
        color: defaultTheme.palette.text.secondary,
        minHeight: 0,
      },
    },
    MuiCardActions: {
      root: {
        justifyContent: 'center',
      },
    },
    MuiInputLabel: {
      formControl: {
        transform: 'none',
      },
    },
  },
  shape: {
    borderRadius: 2,
  },
});
