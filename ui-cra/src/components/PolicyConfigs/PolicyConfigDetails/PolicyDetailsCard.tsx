import { Card, CardContent } from '@material-ui/core';
import { formatURL, Link } from '@weaveworks/weave-gitops';
import { GetPolicyConfigResponse } from '../../../cluster-services/cluster_services.pb';
import { Routes } from '../../../utils/nav';
import {
  PolicyDetailsCardWrapper,
  usePolicyConfigStyle,
  WarningIcon,
} from '../PolicyConfigStyles';

export const renderParameterValue = (param: any) => {
  if (Array.isArray(param)) return param.join(', ');
  else {
    const paramType = typeof param;
    switch (paramType) {
      case 'boolean':
        return paramType ? 'True' : 'False';
      default:
        return param;
    }
  }
};

function PolicyDetailsCard({
  policies,
  totalPolicies,
  clusterName,
}: GetPolicyConfigResponse) {
  const classes = usePolicyConfigStyle();

  return (
    <div>
      <label className={classes.sectionTitle}>
        Policies <span data-testid="totalPolicies">({totalPolicies})</span>
      </label>
      <PolicyDetailsCardWrapper role="list">
        {policies?.map(policy => (
          <li key={policy.id} role="list-item">
            <Card>
              <CardContent>
                {policy.status === 'OK' ? (
                  <Link
                    to={formatURL(Routes.PolicyDetails, {
                      clusterName: clusterName,
                      id: policy.id,
                    })}
                    className={classes.link}
                    data-policy-name={policy.name}
                  >
                    <span data-testid={`policyId-${policy.name}`}>
                      {policy.name}
                    </span>
                  </Link>
                ) : (
                  <div className={classes.policyTitle}>
                    <span
                      title={`One or more policies are not found in cluster ${clusterName}.`}
                      data-testid={`warning-icon-${policy.id}`}
                    >
                      <WarningIcon />
                    </span>
                    <span data-testid={`policyId-${policy.id}`}>
                      {policy.id}
                    </span>
                  </div>
                )}
                <label className="cardLbl">Parameters</label>
                {Object.entries(policy.parameters || {}).map(param => (
                  <div className="parameterItem" key={param[0]}>
                    <label data-testid={param[0]} className={classes.upperCase}>
                      {param[0]}{' '}
                    </label>
                    <div
                      data-testid={`${param[0]}Value`}
                      className={`parameterItemValue ${classes.upperCase}`}
                    >
                      {renderParameterValue(param[1])}
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          </li>
        ))}
      </PolicyDetailsCardWrapper>
    </div>
  );
}

export default PolicyDetailsCard;
