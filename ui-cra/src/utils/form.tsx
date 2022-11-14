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
  FormHelperText as MuiFormHelperText,
} from '@material-ui/core';
import { InputBaseProps } from '@material-ui/core/InputBase';
import { Theme, withStyles } from '@material-ui/core/styles';
import React, { Dispatch, FC } from 'react';

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

const FormHelperText = withStyles(() => ({
  error: {
    color: '#9F3119',
  },
}))(MuiFormHelperText);

const InputBase = withStyles(() => ({
  inputMultiline: {
    padding: '10px',
  },
  error: {
    borderBottom: '2px solid #9F3119',
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
>;

interface InputProps extends PickedInputProps {
  label?: string;
  className?: string;
  multiline?: boolean;
  rows?: number;
  description?: string;
  required?: boolean;
  name?: string;
  error?: boolean;
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
  description,
  required,
  name,
  error,
}) => {
  return (
    <FormControl id={`${label}-group`} className={className}>
      {label && (
        <InputLabel htmlFor={`${label}-input`} shrink>
          {label}
        </InputLabel>
      )}
      <InputBase
        className={className}
        name={name}
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
        required={required}
        error={error}
      />
      <FormHelperText error={error}>
        {!error ? description : 'Please fill this field in.'}
      </FormHelperText>
    </FormControl>
  );
};

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

export const validateFormData = (
  event: any,
  onSubmit: any,
  setFormError: Dispatch<React.SetStateAction<any>>,
) => {
  const requiredButEmptyInputs = Array.from(
    document.forms['form' as unknown as number].querySelectorAll(
      'input[required]',
    ),
  ).filter(input => (input as HTMLInputElement).value === '');

  if (requiredButEmptyInputs.length === 0) {
    onSubmit();
  } else {
    (requiredButEmptyInputs[0] as HTMLInputElement).focus();
    setFormError((requiredButEmptyInputs[0] as HTMLInputElement).name);
  }
};
