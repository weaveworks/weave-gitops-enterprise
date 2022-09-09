import {
  Button,
  Divider as MuiDivider,
  FormControl as MuiFormControl,
  InputLabel as MuiInputLabel,
  MenuItem,
  Typography,
  Select as MuiSelect,
  SelectProps as MuiSelectProps,
  InputBase as MuiInputBase,
  Input as MuiInput,
  FormHelperText,
} from '@material-ui/core';
import { InputBaseProps } from '@material-ui/core/InputBase';
import { Theme, withStyles } from '@material-ui/core/styles';
import React, { FC } from 'react';

// FIXME: what sure what the type should be to export correctly!
export const SectionTitle: any = withStyles(() => ({
  root: {
    margin: '12px 0 8px 0',
  },
}))(Typography);

// FIXME: what sure what the type should be to export correctly!
export const RemoveGroupSectionButton: any = withStyles((theme: Theme) => ({
  root: {
    marginLeft: theme.spacing(0.5),
    fontSize: '12px',
    color: theme.palette.primary.main,
  },
}))(Button);

// FIXME: what sure what the type should be to export correctly!
export const AddGroupSectionButton: any = withStyles((theme: Theme) => ({
  root: {
    marginTop: theme.spacing(1.5),
  },
}))(Button);

const FormControl = withStyles((theme: Theme) => ({
  root: {
    paddingBottom: '24px',
  },
}))(MuiFormControl);

const InputLabel = withStyles(() => ({
  root: {
    fontSize: 12,
    color: 'black',
    paddingBottom: 6,
  },
  formControl: {
    position: 'initial',
  },
}))(MuiInputLabel);

export const InputBase = withStyles(() => ({
  inputMultiline: {
    padding: '10px',
  },
}))(MuiInputBase);

type PickedInputProps = Pick<
  InputBaseProps,
  | 'autoFocus'
  | 'onChange'
  | 'value'
  | 'disabled'
  | 'defaultValue'
  | 'type'
  | 'disabled'
  | 'fullWidth'
  | 'placeholder'
  | 'inputProps'
>;

interface InputProps extends PickedInputProps {
  disabled?: boolean;
  label?: string;
  className?: string;
  multiline?: boolean;
  rows?: number;
  description?: string;
  required?: boolean;
  name?: string;
  InputLabelProps?: any;
  InputProps?: any;
}

export const Input: FC<InputProps> = ({
  label,
  className,
  description,
  disabled,
  fullWidth,
  name,
  inputProps,
  InputProps,
  InputLabelProps,
  children,
  ...rest
}) => (
  <FormControl
    id={`${label}-group`}
    fullWidth={fullWidth}
    disabled={disabled}
    className={className}
  >
    {label && (
      <InputLabel htmlFor={`${label}-input`} shrink {...InputLabelProps}>
        {label}
      </InputLabel>
    )}
    {children || (
      <InputBase
        name={name}
        inputProps={{ ...inputProps, maxLength: 256 }}
        {...InputProps}
        {...rest}
      />
    )}
    <FormHelperText>{description}</FormHelperText>
  </FormControl>
);

// FIXME: what sure what the type should be to export correctly!
export const Divider: any = withStyles((theme: Theme) => ({
  root: {
    marginLeft: theme.spacing(0.5),
    marginRight: 0,
    flexGrow: 1,
  },
}))(MuiDivider);

export const DividerWrapper: FC = ({ children }) => (
  <div style={{ display: 'flex', alignItems: 'center', minHeight: 35 }}>
    {children}
  </div>
);

interface SelectProps extends MuiSelectProps {
  label: string;
  items?: string[];
  value: string;
  disabled?: boolean;
  className?: string;
  description?: string;
}

export const Select: FC<SelectProps> = ({
  children,
  label,
  input,
  items,
  value,
  variant,
  onChange,
  className,
  description,
}) => (
  <FormControl id={`${label}-group`} className={className}>
    <InputLabel htmlFor={`${label}-input`} shrink>
      {label}
    </InputLabel>
    <MuiSelect
      id={`${label}-input`}
      input={input ?? <InputBase />}
      onChange={onChange}
      value={value}
      variant={variant ?? 'outlined'}
    >
      {children ??
        items?.map(item => (
          <MenuItem key={item} value={item}>
            {item}
          </MenuItem>
        ))}
    </MuiSelect>
    <FormHelperText>{description}</FormHelperText>
  </FormControl>
);

export const validateFormData = (event: any, onSubmit: any) => {
  const { form } = event.currentTarget;
  const isValid = form?.reportValidity();
  event.preventDefault();
  if (isValid) {
    onSubmit();
  } else {
    const invalid: HTMLElement | null = form.querySelector(':invalid');
    if (invalid) {
      invalid.focus();
    }
  }
};
