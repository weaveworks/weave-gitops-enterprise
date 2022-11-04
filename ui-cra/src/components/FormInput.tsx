import {
  FormControl,
  FormLabel,
  TextField,
  TextFieldProps,
} from '@material-ui/core';
import styled from 'styled-components';
import { FormContext } from './ControlledForm';

export type FormInputProps = {
  name: string;
  className?: string;
  label: string;
  component?: any;
  valuePropName?: string;
} & TextFieldProps;

function FormInput({
  className,
  label,
  name,
  component,
  valuePropName = 'value',
  ...props
}: FormInputProps) {
  const Component = component || TextField;

  return (
    <FormContext.Consumer>
      {({ handleChange, findValue }) => (
        <FormControl className={className}>
          <FormLabel>{label}</FormLabel>
          <Component
            id={name}
            variant="outlined"
            {...props}
            {...{ [valuePropName]: findValue(name) }}
            onChange={(ev: any) => {
              const v = ev.target[valuePropName];
              handleChange(name, v);
            }}
          />
        </FormControl>
      )}
    </FormContext.Consumer>
  );
}

export default styled(FormInput).attrs({ className: 'FormInput' })`
  min-width: 300px;
`;
