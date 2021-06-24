import { Theme } from '@material-ui/core';
import { createMuiTheme } from '@material-ui/core/styles';

const boxShadow = '0 1px 3px rgba(0,0,0,0.12), 0 1px 2px rgba(0,0,0,0.24)';

const defaultTheme = createMuiTheme();

export const muiTheme = createMuiTheme({
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
    },
    MuiTableHead: {
      root: {
        backgroundColor: '#f9f9fb',
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
  palette: {
    primary: {
      '50': '#00A7CC',
      '100': '#00A7CC',
      '200': '#00A7CC',
      '300': '#00A7CC',
      '400': '#00A7CC',
      '500': '#00A7CC',
      '600': '#00A7CC',
      '700': '#00A7CC',
      '800': '#00A7CC',
      '900': '#00A7CC',
      A100: '#00A7CC',
      A200: '#00A7CC',
      A400: '#00A7CC',
      A700: '#00A7CC',
    },
  },
  shape: {
    borderRadius: 2,
  },
  typography: {
    fontFamily: ['proxima-nova', 'Helvetica', 'Arial', 'sans-serif'].join(', '),
  },
});

export const popperArrow = (theme: Theme) => ({
  '&[x-placement*="bottom"] $arrow': {
    '&::before': {
      borderColor: `transparent transparent ${theme.palette.background.paper} transparent`,
      borderWidth: '0 1em 1em 1em',
    },
    height: '1em',
    left: 0,
    marginTop: '-0.9em',
    top: 0,
    width: '3em',
  },
  '&[x-placement*="left"] $arrow': {
    '&::before': {
      borderColor: `transparent transparent transparent ${theme.palette.background.paper}`,
      borderWidth: '1em 0 1em 1em',
    },
    height: '3em',
    marginRight: '-0.9em',
    right: 0,
    width: '1em',
  },
  '&[x-placement*="right"] $arrow': {
    '&::before': {
      borderColor: `transparent ${theme.palette.background.paper} transparent transparent`,
      borderWidth: '1em 1em 1em 0',
    },
    height: '3em',
    left: 0,
    marginLeft: '-0.9em',
    width: '1em',
  },
  '&[x-placement*="top"] $arrow': {
    '&::before': {
      borderColor: `${theme.palette.background.paper} transparent transparent transparent`,
      borderWidth: '1em 1em 0 1em',
    },
    bottom: 0,
    height: '1em',
    left: 0,
    marginBottom: '-0.9em',
    width: '3em',
  },
  zIndex: 1,
});
