import { createTheme } from '@material-ui/core/styles';
import { muiTheme as coreMuiTheme } from '@weaveworks/weave-gitops';

const boxShadow = '0 1px 3px rgba(0,0,0,0.12), 0 1px 2px rgba(0,0,0,0.24)';

const defaultTheme = createTheme();

export const muiTheme = createTheme({
  ...coreMuiTheme,
  overrides: {
    MuiButton: {
      contained: {
        backgroundColor: 'hsl(0, 0%, 100%)',
        color: 'hsl(0, 0%, 45%)',
        boxShadow,
        '&:hover': {
          backgroundColor: 'hsl(0, 0%, 96%)',
          boxShadow,
          color: 'hsl(240, 20%, 30%)',
        },
      },
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
        backgroundColor: '#FFFFFF',
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
        border: '1px solid #E5E5E5',
        borderRadius: 2,
        position: 'relative',
        backgroundColor: defaultTheme.palette.common.white,
        fontSize: 16,
        width: '100%',
        padding: '8px 12px',
        '&:focus': {
          borderColor: '#00A7CC',
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
        color: '#888888',
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
  // palette: {
  //   primary: {
  //     '500': '#00A7CC',
  //   },
  // },
  shape: {
    borderRadius: 2,
  },
  // typography: {
  //   fontFamily: ['proxima-nova', 'Helvetica', 'Arial', 'sans-serif'].join(', '),
  // },
});
