import { CircularProgress } from '@material-ui/core';
import { Flex } from '@weaveworks/weave-gitops';
import Alert from '@material-ui/lab/Alert';
import { useGetGithubDeviceCode } from '../../contexts/GitAuth';
import { storeProviderToken } from './utils';
import Modal from '../Modal';
import ModalContent from './ModalContent';
import { GitProvider } from '../../api/gitauth/gitauth.pb';

type Props = {
  className?: string;
  bodyClassName?: string;
  open: boolean;
  onSuccess: (token: string) => void;
  onClose: () => void;
  repoName: string;
};

export function GithubDeviceAuthModal({
  className,
  bodyClassName,
  open,
  onClose,
  repoName,
  onSuccess,
}: Props) {
  const { isLoading, error, data } = useGetGithubDeviceCode();
  return (
    <Modal
      className={className}
      bodyClassName={bodyClassName}
      title="Authenticate with Github"
      open={open}
      onClose={onClose}
      description={`Weave GitOps needs to authenticate with the Git Provider for the ${repoName} repo`}
    >
      <p>
        Paste this code into the Github Device Activation field to grant Weave
        GitOps temporary access:
      </p>
      {error && (
        <Alert severity="error" title="Error">
          {error.message}
        </Alert>
      )}
      <Flex wide center height="150px">
        {isLoading || !data ? (
          <CircularProgress />
        ) : (
          <ModalContent
            onSuccess={(token: string) => {
              storeProviderToken(GitProvider.GitHub, token);
              onSuccess(token);
              onClose();
            }}
            codeRes={data}
          />
        )}
      </Flex>
    </Modal>
  );
}
