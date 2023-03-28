import styled from 'styled-components';
import { Body, Title } from '../Shared';
import { Flex, Link, MessageBox, YamlView } from '@weaveworks/weave-gitops';

type Props = {
  plan?: string;
  error?: string;
};

function TerraformPlanView({ plan, error }: Props) {
  return (
    <Flex align wide tall column>
      {plan && !error ? (
        <YamlView yaml={plan.trimStart() || ''} />
      ) : (
        <MessageBox>
          <Title>Terraform Plan</Title>
          <Body>No plan available.</Body>
          <Body>
            To enable the plan view, please set the field
            `spec.storeReadablePlan` to `human`.
          </Body>
          <Body>
            To learn more about planning Terraform resources,&nbsp;
            <Link
              href="https://docs.gitops.weave.works/docs/terraform/Using%20Terraform%20CR/plan-and-manually-apply-terraform-resources/"
              newTab
            >
              visit our documentation
            </Link>
          </Body>
        </MessageBox>
      )}
    </Flex>
  );
}

export default styled(TerraformPlanView).attrs({
  className: TerraformPlanView.name,
})``;
