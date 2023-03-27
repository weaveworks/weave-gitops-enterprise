import { useCallback, useEffect, useState } from 'react';
import { Input } from '../../../utils/form';
import { useDebounce } from '../../../utils/hooks';

const InputDebounced = ({
  required = true,
  name,
  label,
  placeholder = '',
  value,
  handleFormData,
}: {
  required: boolean;
  name: string;
  placeholder?: string;
  label: string;
  value: string;
  handleFormData: (value: any) => void;
}) => {
  const [data, setData] = useState<string>(value);
  const debouncedValue = useDebounce<string>(data);
  useEffect(() => {
    handleFormData(debouncedValue);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedValue]);

  return (
    <Input
      className="form-section"
      required={required}
      name={name}
      label={label}
      placeholder={placeholder}
      value={data}
      onChange={e => setData(e.target.value)}
    />
  );
};

export default InputDebounced;
