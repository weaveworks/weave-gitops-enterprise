import styled from 'styled-components';
import { GetTerraformObjectPlanResponse } from '../../api/terraform/terraform.pb';
import { useGetTerraformObjectPlan } from '../../contexts/Terraform';
import CodeView from '../CodeView';
import { Body, Message, Title } from '../Shared';

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function TerraformPlanView({ className, ...params }: Props) {
  const { data, isLoading } = useGetTerraformObjectPlan(params);
  const { plan } = (data || {}) as GetTerraformObjectPlanResponse;

  if (isLoading) {
    return <></>;
  }

  return (
    <>
      {plan ? (
        <CodeView
          kind="Terraform"
          code={plan.trimStart() || ''}
          colorizeChanges
        />
      ) : (
        <Message>
          <Title>Terraform Plan</Title>
          <Body>No plan available.</Body>
        </Message>
      )}
    </>
  );
}

export default styled(TerraformPlanView).attrs({
  className: TerraformPlanView.name,
})``;
