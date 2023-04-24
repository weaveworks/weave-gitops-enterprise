import styled from 'styled-components';
import { Flex, Link, MessageBox, Text, Spacer } from '@weaveworks/weave-gitops';
import CodeView from '../CodeView';

type Props = {
  plan?: string;
  error?: string;
};

function TerraformPlanView({ plan, error }: Props) {
  return (
    <Flex align wide tall column>
      {plan && !error ? (
        <CodeView
          kind="Terraform"
          code={plan.trimStart() || ''}
          colorizeChanges
        />
      ) : (
        <MessageBox>
          <Spacer padding="small" />
          <Text size="large" semiBold>
            Terraform Plan
          </Text>
          <Spacer padding="small" />
          <Text size="medium">No plan available.</Text>
          <Spacer padding="small" />
          <Text size="medium">
            To enable the plan view, please set the field
            `spec.storeReadablePlan` to `human`.
          </Text>
          <Spacer padding="small" />
          <Text size="medium">
            To learn more about planning Terraform resources,&nbsp;
            <Link
              href="https://docs.gitops.weave.works/docs/terraform/Using%20Terraform%20CR/plan-and-manually-apply-terraform-resources/"
              newTab
            >
              visit our documentation.
            </Link>
          </Text>
        </MessageBox>
      )}
    </Flex>
  );
}

export default styled(TerraformPlanView).attrs({
  className: TerraformPlanView.name,
})``;
