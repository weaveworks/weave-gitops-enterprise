import styled from 'styled-components';
import {
  Flex,
  Link,
  MessageBox,
  YamlView,
  Text,
  Spacer,
} from '@weaveworks/weave-gitops';

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
