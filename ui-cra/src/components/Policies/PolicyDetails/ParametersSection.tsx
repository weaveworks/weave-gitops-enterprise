import {
  Policy,
  PolicyParam,
} from '../../../cluster-services/cluster_services.pb';
import { usePolicyStyle } from '../PolicyStyles';

function ParametersSection({ parameters }: Policy) {
  const classes = usePolicyStyle();
  const parseValue = (parameter: PolicyParam) => {
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
      <div className={classes.sectionSeperator}>
        <div className={classes.cardTitle}>Parameters Definition:</div>
        {parameters?.map((parameter: PolicyParam) => (
          <div key={parameter.name} className={classes.parameterWrapper}>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Parameter Name</span>
              <span className={classes.body1}>{parameter.name}</span>
            </div>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Parameter Type</span>
              <span className={classes.body1}>{parameter.type}</span>
            </div>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Value</span>
              <span className={classes.body1}>
                {parameter.value ? (
                  parseValue(parameter)
                ) : (
                  <div className={classes.chip}>undefined</div>
                )}
              </span>
            </div>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Required</span>
              <span className={classes.body1}>
                {parameter.required ? 'True' : 'False'}
              </span>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}

export default ParametersSection;
