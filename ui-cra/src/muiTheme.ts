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
        marginRight: weaveTheme.spacing.small,
      },
      outlinedPrimary: {
        borderColor: weaveTheme.colors.neutral20,
        '&:hover': {
          borderColor: weaveTheme.colors.neutral20,
        },
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

export const localEEMuiTheme = createTheme({
  ...muiTheme,
  overrides: {
    ...muiTheme.overrides,
    MuiInputBase: {
      ...muiTheme.overrides?.MuiInputBase,
      root: {
        ...muiTheme.overrides?.MuiInputBase?.root,
        marginRight: `${weaveTheme.spacing.medium}`,
      },
      input: {
        ...muiTheme.overrides?.MuiInputBase?.input,
        minWidth: '155px',
        border: `1px solid ${weaveTheme.colors.neutral20}`,
        position: 'relative',
        fontSize: 16,
        width: '100%',
        padding: '8px 12px',
        '&:focus': {
          borderColor: weaveTheme.colors.primaryDark,
          borderRadius: 2,
        },
      },
    },
    MuiInputLabel: {
      ...muiTheme.overrides?.MuiInputLabel,
      formControl: {
        ...muiTheme.overrides?.MuiInputLabel?.formControl,
        fontSize: `${weaveTheme.fontSizes.tiny}`,
      },
      shrink: {
        transform: 'none',
      },
      asterisk: {
        display: 'none',
      },
    },
    MuiSelect: {
      select: {
        ...muiTheme.overrides?.MuiSelect?.select,
        minWidth: '155px',
      },
    },
    MuiCheckbox: {
      root: {
        ...muiTheme.overrides?.MuiCheckbox?.root,
        padding: 0,
      },
    },
  },
});
