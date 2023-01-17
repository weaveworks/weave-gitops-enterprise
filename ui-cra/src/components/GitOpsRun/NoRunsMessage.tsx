import { Link } from '@weaveworks/weave-gitops';
import { Body, Message, Title } from '../Shared';

const NoRunsMessage = () => {
  return (
    <Message>
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
    </Message>
  );
};

export default NoRunsMessage;
