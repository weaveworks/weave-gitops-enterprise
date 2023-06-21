import { Card, CardContent } from '@material-ui/core';
import { Link, formatURL } from '@weaveworks/weave-gitops';
import {
  GetPolicyConfigResponse,
  PolicyConfigPolicy,
} from '../../../cluster-services/cluster_services.pb';
import { Routes } from '../../../utils/nav';
import {
  PolicyDetailsCardWrapper,
  SectionTitle,
  WarningIcon,
  usePolicyConfigStyle,
} from '../PolicyConfigStyles';

interface GetCardTitleProps {
  policy: PolicyConfigPolicy;
  clusterName: string;
}

export const renderParameterValue = (param: any) => {
  if (Array.isArray(param)) return param.join(', ');
  const paramType = typeof param;
  switch (paramType) {
    case 'boolean':
      return paramType ? 'True' : 'False';
    default:
      return param;
  }
};

export default function PolicyDetailsCard({
  policies,
  totalPolicies,
  clusterName,
}: GetPolicyConfigResponse) {
  const classes = usePolicyConfigStyle();

  return (
    <div>
      <SectionTitle>
        Policies <span data-testid="totalPolicies">({totalPolicies})</span>
      </SectionTitle>
      <PolicyDetailsCardWrapper>
        {policies?.map(policy => (
          <li key={policy.id} data-testid="list-item">
            <Card>
              <CardContent>
                <CardTitle clusterName={clusterName || ''} policy={policy} />
                <label className="cardLbl">Parameters</label>
                {Object.entries(policy.parameters || {}).map(([key, value]) => (
                  <div className="parameterItem" key={key}>
                    <label data-testid={key} className={classes.upperCase}>
                      {key}{' '}
                    </label>
                    <div
                      data-testid={`${key}Value`}
                      className={`parameterItemValue ${classes.upperCase}`}
                    >
                      {renderParameterValue(value)}
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
export function CardTitle({ clusterName, policy }: GetCardTitleProps) {
  const classes = usePolicyConfigStyle();
  const { status, id, name } = policy;

  return status === 'OK' ? (
    <Link
      to={formatURL(Routes.PolicyDetails, {
        clusterName: clusterName,
        id: id,
      })}
      className={classes.link}
      data-policy-name={name}
    >
      <span data-testid={`policyId-${name}`}>{name}</span>
    </Link>
  ) : (
    <div className={classes.policyTitle}>
      <span
        title={`One or more policies are not found in cluster ${clusterName}.`}
        data-testid={`warning-icon-${id}`}
      >
        <WarningIcon />
      </span>
      <span data-testid={`policyId-${id}`}>{id}</span>
    </div>
  );
}
