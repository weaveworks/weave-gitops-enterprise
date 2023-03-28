import { Link, MessageBox } from '@weaveworks/weave-gitops';
import { Body, Title } from '../../Shared';

const NoRunsMessage = () => {
  return (
    <MessageBox>
      <Title>GitOps Run</Title>
      <Body>
        There are currently no instances of GitOps run running on the cluster.
      </Body>
      <Body>
        To learn more about GitOps run and how to get started, check out our
        documentation{' '}
        <Link
          href="https://docs.gitops.weave.works/docs/gitops-run/overview/"
          newTab
        >
          here
        </Link>
      </Body>
    </MessageBox>
  );
};

export default NoRunsMessage;
