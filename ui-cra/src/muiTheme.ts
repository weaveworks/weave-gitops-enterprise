import { Theme, createTheme } from '@material-ui/core/styles';
import { baseTheme, muiTheme as coreMuiTheme } from '@weaveworks/weave-gitops';
import { ThemeTypes } from '@weaveworks/weave-gitops/ui/contexts/AppContext';

const defaultTheme = createTheme();

export const muiTheme = (colors: any, mode: ThemeTypes) =>
  createTheme({
    ...coreMuiTheme(colors, mode),
    overrides: {
      MuiButton: {
        root: {
          textTransform: 'none',
          minWidth: 52,
          marginRight: baseTheme.spacing.small,
        },
        outlinedPrimary: {
          borderColor: `${colors.neutral20} !important`,
          '&:hover': {
            borderColor: colors.neutral20,
          },
        },
      },
      MuiDialog: {
        root: {
          padding: 0,
        },
        paper: {
          padding: 0,
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
          color: colors.neutral30,
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

export const localEEMuiTheme = (theme: Theme) =>
  createTheme({
    ...theme,
    overrides: {
      ...theme.overrides,
      MuiInputBase: {
        ...theme.overrides?.MuiInputBase,
        root: {
          ...theme.overrides?.MuiInputBase?.root,
          marginRight: `${baseTheme.spacing.medium}`,
        },
        input: {
          ...theme.overrides?.MuiInputBase?.input,
          minWidth: '155px',
          // border: `1px solid ${baseTheme.colors.neutral20}`,
          position: 'relative',
          fontSize: 16,
          width: '100%',
          padding: '8px 12px',
          '&:focus': {
            borderRadius: 2,
          },
        },
      },
      MuiInputLabel: {
        ...theme.overrides?.MuiInputLabel,
        formControl: {
          ...theme.overrides?.MuiInputLabel?.formControl,
          fontSize: `${baseTheme.fontSizes.tiny}`,
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
          ...theme.overrides?.MuiSelect?.select,
          minWidth: '155px',
        },
      },
      MuiCheckbox: {
        root: {
          ...theme.overrides?.MuiCheckbox?.root,
          padding: 0,
        },
      },
    },
  });
