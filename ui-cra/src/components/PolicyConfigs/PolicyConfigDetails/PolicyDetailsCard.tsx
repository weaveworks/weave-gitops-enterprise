import { Card, CardContent } from '@material-ui/core';
import { formatURL, Link } from '@weaveworks/weave-gitops';
import { GetPolicyConfigResponse } from '../../../cluster-services/cluster_services.pb';
import { Routes } from '../../../utils/nav';
import {
    PolicyDetailsCardWrapper,
    usePolicyConfigStyle,
    WarningIcon
} from '../PolicyConfigStyles';

function PolicyDetailsCard({
  policies,
  totalPolicies,
  clusterName,
}: GetPolicyConfigResponse) {
  const classes = usePolicyConfigStyle();

  return (
    <div>
      <label className={classes.sectionTitle}>
        Policies <span>({totalPolicies})</span>
      </label>
      <PolicyDetailsCardWrapper>
        {policies?.map(policy => (
          <li key={policy.id}>
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
                    {policy.name}
                  </Link>
                ) : (
                  <div>
                    <span
                      title={`One or more policies are not found in cluster ${clusterName}.`}
                      data-testid={`warning-icon-${policy.name}`}
                    >
                      <WarningIcon />
                    </span>
                    {policy.name}
                  </div>
                )}
                <label className="cardLbl">Parameters</label>
                {Object.entries(policy.parameters || {}).map(param => (
                  <div className="parameterItem">
                    <label>{param[0]}: </label>
                    <div className="parameterItemValue">{param[1]}</div>
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
