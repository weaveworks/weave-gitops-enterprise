import { FC, useEffect, useState } from 'react';
import { Input, InputProps } from '../../../utils/form';
import { useDebounce } from '../../../utils/hooks';

interface InputDebounceProps extends InputProps {
  value?: string;
  handleFormData: (value: any) => void;
}

const InputDebounced: FC<InputDebounceProps> = ({
  value,
  handleFormData,
  ...rest
}) => {
  const [data, setData] = useState<string>(value || '');
  const [error, setError] = useState<boolean>(false);
  const debouncedValue = useDebounce<string>(data);

  const handleChange = (e: any) => {
    setData(e.target.value);
    setError(!e.target.value);
  };

  const handleBlur = () => {
    setError(!data && (rest.required || false));
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
      error={error}
    />
  );
};

export default InputDebounced;
