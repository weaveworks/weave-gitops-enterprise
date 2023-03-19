import { useCallback, useState } from 'react';
import { Input } from '../../../utils/form';

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
  const debounce = () => {
    let timer: NodeJS.Timeout | null;
    return function (val: string) {
      setData(val);
      if (timer) clearTimeout(timer);
      timer = setTimeout(() => {
        timer = null;
        handleFormData(val);
      }, 500);
    };
  };
  const handleOnChange = useCallback(debounce(), []);

  return (
    <Input
      className="form-section"
      required={required}
      name={name}
      label={label}
      placeholder={placeholder}
      value={data}
      onChange={e => handleOnChange(e.target.value)}
    />
  );
};

export default InputDebounced;
