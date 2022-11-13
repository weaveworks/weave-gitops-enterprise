import { PolicyValidationParam } from '../../../cluster-services/cluster_services.pb';
import {
  ParameterCell,
  parseValue,
} from '../../Policies/PolicyDetails/PolicyUtilis';
import { ParameterWrapper, usePolicyStyle } from '../../Policies/PolicyStyles';
interface IViolationDetailsProps {
  parameters: PolicyValidationParam[] | undefined;
}

function PolicyConfigSection({ parameters }: IViolationDetailsProps) {
  const classes = usePolicyStyle();

  return (
    <>
      <div className={classes.cardTitle}>Parameters Values:</div>
      {parameters?.map(parameter => (
        <ParameterWrapper key={parameter.name} id={parameter.name}>
          <ParameterCell
            label="Parameter Name"
            value={parameter.name}
          ></ParameterCell>
          <ParameterCell
            label="Value"
            value={parseValue(parameter)}
          ></ParameterCell>
          <ParameterCell
            label="Policy Config Name"
            value={parameter.configRef || '-'}
          ></ParameterCell>
        </ParameterWrapper>
      ))}
    </>
  );
}

export default PolicyConfigSection;
