import { Card, CardContent } from '@material-ui/core';
import {
  Flex,
  Link,
  Text,
  V2Routes,
  formatURL,
} from '@weaveworks/weave-gitops';
import {
  GetPolicyConfigResponse,
  PolicyConfigPolicy,
} from '../../../cluster-services/cluster_services.pb';
import {
  PolicyDetailsCardWrapper,
  SectionTitle,
  WarningIcon,
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
                    <Text data-testid={key} uppercase>
                      {key}
                    </Text>
                    <Text
                      uppercase
                      className="parameterItemValue"
                      data-testid={`${key}Value`}
                    >
                      {renderParameterValue(value)}
                    </Text>
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
  const { status, id, name } = policy;

  return status === 'OK' ? (
    <Link
      textProps={{
        color: 'primary',
        size: 'medium',
        capitalize: true,
      }}
      to={formatURL(V2Routes.PolicyDetailsPage, {
        clusterName: clusterName,
        id: id,
      })}
      data-policy-name={name}
    >
      <span data-testid={`policyId-${name}`}>{name}</span>
    </Link>
  ) : (
    <Flex align gap="8">
      <span
        title={`One or more policies are not found in cluster ${clusterName}.`}
        data-testid={`warning-icon-${id}`}
      >
        <WarningIcon />
      </span>
      <span data-testid={`policyId-${id}`}>{id}</span>
    </Flex>
  );
}
