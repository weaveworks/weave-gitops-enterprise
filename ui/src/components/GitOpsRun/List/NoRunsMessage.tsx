import { Link, MessageBox, Spacer, Text } from '@weaveworks/weave-gitops';

const NoRunsMessage = () => {
  return (
    <MessageBox>
      <Spacer padding="small" />
      <Text size="large" semiBold>
        GitOps Run
      </Text>
      <Spacer padding="small" />
      <Text size="medium">
        There are currently no instances of GitOps run running on the cluster.
      </Text>
      <Spacer padding="small" />
      <Text size="medium">
        To learn more about GitOps run and how to get started, check out our
        documentation{' '}
        <Link
          href="https://docs.gitops.weave.works/docs/gitops-run/overview/"
          newTab
        >
          here
        </Link>
      </Text>
    </MessageBox>
  );
};

export default NoRunsMessage;
