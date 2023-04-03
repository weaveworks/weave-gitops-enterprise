import { ChipWrapper, ParameterInfo } from '../PolicyStyles';

export const parseValue = (parameter: {
  type?: string | undefined;
  value?: any;
}) => {
  if (!parameter.value) return <ChipWrapper>undefined</ChipWrapper>;
  switch (parameter.type) {
    case 'boolean':
      return parameter.value.value ? 'true' : 'false';
    case 'array':
      return parameter.value.value.join(', ');
    case 'string':
      return parameter.value.value;
    case 'integer':
      return parameter.value.value.toString();
    default:
  }
};

export const ParameterCell = ({
  label,
  value,
}: {
  label: string;
  value: string | undefined;
}) => {
  return (
    <ParameterInfo column data-testid={label}>
      <span className="label">{label}</span>
      <span className="body1">{value}</span>
    </ParameterInfo>
  );
};
