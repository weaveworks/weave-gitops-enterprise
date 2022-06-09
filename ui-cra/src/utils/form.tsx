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
    display: 'flex',
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: theme.spacing(0.75),
    marginTop: theme.spacing(0.75),
  },
}))(MuiFormControl);

const InputLabel = withStyles(() => ({
  root: {
    width: 125,
    fontSize: 16,
    color: 'black',
  },
  formControl: {
    position: 'initial',
  },
}))(MuiInputLabel);

const InputBase = withStyles(() => ({
  input: {
    marginLeft: '10px',
  },
  inputMultiline: {
    padding: '10px',
  },
}))(MuiInputBase);

type PickedInputProps = Pick<
  InputBaseProps,
  | 'autoFocus'
  | 'onChange'
  | 'value'
  | 'defaultValue'
  | 'type'
  | 'disabled'
  | 'placeholder'
  | 'error'
>;

interface InputProps extends PickedInputProps {
  label?: string;
  className?: string;
  multiline?: boolean;
  rows?: number;
}

export const Input: FC<InputProps> = ({
  autoFocus,
  defaultValue,
  disabled,
  value,
  onChange,
  label,
  type,
  placeholder,
  className,
  multiline,
  rows,
  error,
}) => (
  <FormControl className={className}>
    {label && (
      <InputLabel htmlFor={`${label}-input`} shrink>
        {label}
      </InputLabel>
    )}
    <InputBase
      autoFocus={autoFocus}
      defaultValue={defaultValue}
      disabled={disabled}
      id={`${label}-input`}
      onChange={onChange}
      placeholder={placeholder}
      type={type}
      value={value}
      multiline={multiline}
      rows={rows}
      inputProps={{ maxLength: 256 }}
      error={error}
    />
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
}

export const Select: FC<SelectProps> = ({
  children,
  label,
  input,
  items,
  value,
  variant,
  onChange,
}) => (
  <FormControl>
    <InputLabel htmlFor={`${label}-input`} shrink>
      {label}
    </InputLabel>
    <MuiSelect
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
  </FormControl>
);
