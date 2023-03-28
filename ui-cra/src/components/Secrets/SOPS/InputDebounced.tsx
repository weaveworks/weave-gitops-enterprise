import { FC, useEffect, useState } from 'react';
import { Input, InputProps } from '../../../utils/form';
import { useDebounce } from '../../../utils/hooks';

interface InputDebounceProps extends InputProps {
  value?: string;
  handleFormData: (value: any) => void;
}

const InputDebounced: FC<InputDebounceProps> = ({
  value,
  error,
  handleFormData,
  ...rest
}) => {
  const [data, setData] = useState<string>(value || '');
  const [inputError, setInputError] = useState<boolean>(error || false);
  const debouncedValue = useDebounce<string>(data);

  const handleChange = (e: any) => {
    setData(e.target.value);
  };

  const handleBlur = () => {
    setInputError(!data && (rest.required || false));
  };

  useEffect(() => {
    handleFormData(debouncedValue);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedValue]);

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

export default InputDebounced;
