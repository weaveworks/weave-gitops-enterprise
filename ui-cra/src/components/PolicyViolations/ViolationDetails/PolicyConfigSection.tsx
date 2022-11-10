import {
  PolicyValidationParam,
} from '../../../cluster-services/cluster_services.pb';
import { ChipWrapper, usePolicyStyle } from '../../Policies/PolicyStyles';
interface IViolationDetailsProps {
  parameters: PolicyValidationParam[] | undefined;
}
function PolicyConfigSection({parameters}: IViolationDetailsProps) {
  const classes = usePolicyStyle();
  const parseValue = (parameter: PolicyValidationParam) => {
    switch (parameter.type) {
      case 'boolean':
        return parameter.value.value ? 'true' : 'false';
      case 'array':
        return parameter.value.value.join(', ');
      case 'string':
        return parameter.value.value;
      case 'integer':
        return parameter.value.value.toString();
    }
  };
  return (
    <>
        <div className={classes.cardTitle}>Parameters Values:</div>
        {parameters?.map((parameter) => (
          <div key={parameter.name} className={classes.parameterWrapper}>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Parameter Name</span>
              <span className={classes.body1}>{parameter.name}</span>
            </div>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Value</span>
              <span className={classes.body1}>
                {parameter.value ? (
                  parseValue(parameter)
                ) : (
                  <ChipWrapper> undefined</ChipWrapper>
                )}
              </span>
            </div>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Policy Config Name</span>
              <span className={classes.body1}>
                {parameter.configRef ? parameter.configRef  : '-'}
              </span>
            </div>
          </div>
        ))}
    </>
  );
}

export default PolicyConfigSection;
