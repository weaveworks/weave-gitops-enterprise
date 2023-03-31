import {
  Policy,
  PolicyParam,
} from '../../../cluster-services/cluster_services.pb';
import { ParameterWrapper, usePolicyStyle } from '../PolicyStyles';
import { ParameterCell, parseValue } from './PolicyUtilis';

function ParametersSection({ parameters }: Policy) {
  const classes = usePolicyStyle();
  return (
    <>
      <div className={classes.sectionSeperator}>
        <div className={classes.cardTitle}>Parameters Definition:</div>
        {parameters?.map((parameter: PolicyParam) => (
          <ParameterWrapper key={parameter.name} id={parameter.name}>
            <ParameterCell label="Name" value={parameter.name}></ParameterCell>
            <ParameterCell label="Type" value={parameter.type}></ParameterCell>
            <ParameterCell
              label="Value"
              value={parseValue(parameter)}
            ></ParameterCell>
            <ParameterCell
              label="Required"
              value={parameter.required ? 'True' : 'False'}
            ></ParameterCell>
          </ParameterWrapper>
        ))}
      </div>
    </>
  );
}

export default ParametersSection;
