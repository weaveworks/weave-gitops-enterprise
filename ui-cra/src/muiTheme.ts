import { Theme, createTheme } from '@material-ui/core/styles';
import {
  ThemeTypes,
  baseTheme,
  muiTheme as coreMuiTheme,
} from '@weaveworks/weave-gitops';

const defaultTheme = createTheme();

export const muiTheme = (colors: any, mode: ThemeTypes) => {
  const coreTheme = coreMuiTheme(colors, mode);
  return createTheme({
    ...coreTheme,
    overrides: {
      ...coreTheme.overrides,
      MuiButton: {
        root: {
          textTransform: 'none',
          minWidth: 52,
          marginRight: baseTheme.spacing.small,
          //copied from oss
          '&$disabled': {
            color:
              mode === ThemeTypes.Dark ? colors.primary20 : colors.neutral20,
          },
        },
        outlinedPrimary: {
          borderColor: colors.neutral20,
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
        icon: {
          color: colors.black,
        },
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
        root: {
          color: colors.black,
        },
        formControl: {
          transform: 'none',
        },
      },
    },
    shape: {
      borderRadius: 2,
    },
  });
};

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
