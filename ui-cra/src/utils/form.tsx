import {
  InputAdornment,
  MenuItem,
  Divider as MuiDivider,
  FormControl as MuiFormControl,
  FormHelperText as MuiFormHelperText,
  InputBase as MuiInputBase,
  InputLabel as MuiInputLabel,
  Select as MuiSelect,
  SelectProps as MuiSelectProps,
} from '@material-ui/core';
import { InputBaseProps } from '@material-ui/core/InputBase';
import { Theme, withStyles } from '@material-ui/core/styles';
import { debounce } from 'lodash';
import React, { Dispatch, FC, useEffect, useRef, useState } from 'react';
import styled from 'styled-components';
import { ReactComponent as ErrorIcon } from './../assets/img/error.svg';

const FormControl = withStyles((theme: Theme) => ({
  root: {
    paddingBottom: '24px',
  },
}))(MuiFormControl);

const InputLabel = withStyles(() => ({
  root: {
    fontSize: 14,
    paddingBottom: 6,
  },
  formControl: {
    position: 'initial',
  },
}))(MuiInputLabel);

const InputBase = styled(MuiInputBase)`
  &.MuiInputBase-root {
    border: 2px solid ${props => props.theme.colors.neutralGray};
    border-radius: 2px;
    margin-right: 24px;
    padding: 0px 8px;
  }
  &.Mui-error {
    border-bottom: 2px solid ${props => props.theme.colors.alertDark};
  }
`;

type PickedInputProps = Pick<
  InputBaseProps,
  | 'autoFocus'
  | 'onChange'
  | 'value'
  | 'defaultValue'
  | 'type'
  | 'disabled'
  | 'placeholder'
  | 'onBlur'
>;

export interface InputProps extends PickedInputProps {
  label?: string;
  className?: string;
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
  onBlur,
  label,
  type,
  placeholder,
  className,
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
        name={name}
        autoFocus={autoFocus}
        defaultValue={defaultValue}
        disabled={disabled}
        id={`${label}-input`}
        onChange={onChange}
        onBlur={onBlur}
        placeholder={placeholder}
        type={type}
        value={value}
        rows={rows}
        inputProps={{
          maxLength: 256,
        }}
        endAdornment={
          <InputAdornment position="end" style={{ paddingRight: '12px' }}>
            {error ? <ErrorIcon /> : <></>}
          </InputAdornment>
        }
        required={required}
        error={error}
      />
      <MuiFormHelperText
        style={{
          color: error ? '#9F3119' : '#737373',
        }}
      >
        {!error ? description : 'Please fill this field in.'}
      </MuiFormHelperText>
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

export interface SelectProps extends MuiSelectProps {
  label: string;
  items?: string[];
  value: string;
  disabled?: boolean;
  className?: string;
  description?: string;
  required?: boolean;
  name?: string;
  error?: boolean;
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
  disabled,
  required,
  name,
  error,
}) => (
  <FormControl id={`${label}-group`} className={className}>
    <InputLabel htmlFor={`${label}-input`} shrink>
      {label}
    </InputLabel>
    <MuiSelect
      id={`${label}-input`}
      input={input ?? <InputBase required={required} error={error} />}
      onChange={onChange}
      value={value}
      variant={variant ?? 'outlined'}
      disabled={disabled}
      endAdornment={
        <InputAdornment position="end" style={{ paddingRight: '24px' }}>
          {error ? <ErrorIcon /> : <></>}
        </InputAdornment>
      }
      required={required}
      error={error}
      name={name}
    >
      {children ??
        items?.map(item => (
          <MenuItem key={item} value={item}>
            {item}
          </MenuItem>
        ))}
    </MuiSelect>
    <MuiFormHelperText
      style={{
        color: error ? '#9F3119' : '#737373',
      }}
    >
      {!error ? description : 'Please fill this field in.'}
    </MuiFormHelperText>
  </FormControl>
);

export const validateFormData = (
  event: any,
  onSubmit: any,
  setFormError: Dispatch<React.SetStateAction<any>>,
  setSubmitType?: Dispatch<React.SetStateAction<string>>,
) => {
  event.preventDefault();
  const requiredButEmptyInputs = Array.from(event.target).filter(
    (element: any) =>
      element.type === 'text' && element.required && element.value === '',
  );
  if (requiredButEmptyInputs.length === 0) {
    onSubmit();
  } else {
    const [firstEmpty] = requiredButEmptyInputs;
    (firstEmpty as HTMLInputElement).focus();
    setFormError((firstEmpty as HTMLInputElement).name);
  }
  setSubmitType && setSubmitType('');
};

interface InputDebounceProps extends InputProps {
  value?: string;
  handleFormData: (value: any) => void;
}

export const InputDebounced: FC<InputDebounceProps> = ({
  value,
  error,
  handleFormData,
  ...rest
}) => {
  const [data, setData] = useState<string>(value || '');
  const [inputError, setInputError] = useState<boolean>(error || false);

  const handleBlur = () => {
    setInputError(!data && (rest.required || false));
  };

  const updateFormData = useRef(
    debounce(value => {
      handleFormData(value);
    }, 500),
  ).current;

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { value } = e.target;
    setData(value);
    updateFormData(value);
  };

  useEffect(() => {
    return () => {
      updateFormData.cancel();
    };
  }, [updateFormData]);

  return (
    <Input
      className="form-section"
      {...rest}
      value={data}
      onChange={handleChange}
      onBlur={handleBlur}
      error={error || (inputError && !data)}
    />
  );
};
