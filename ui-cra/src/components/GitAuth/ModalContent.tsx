import styled from 'styled-components';
import {
  Flex,
  Icon,
  IconType,
  Button,
  theme,
  CopyToClipboard,
} from '@weaveworks/weave-gitops';
import { useGetGithubAuthStatus } from '../../contexts/GitAuth';

const { extraLarge } = theme.fontSizes;

const Pad = styled(Flex)`
  padding: 8px 0;
`;

const ModalContent = styled(({ codeRes, onSuccess, className }: any) => {
  const { data } = useGetGithubAuthStatus(codeRes);
  if (!!data) {
    onSuccess(data.accessToken);
  }
  return (
    <div className={className}>
      <Pad wide center>
        <Flex align testId="github-code-container">
          <p className="code-text" data-testid="github-code">
            {codeRes.userCode}
          </p>
          <CopyToClipboard
            value={codeRes.userCode as string}
            className="copy-code"
          />
        </Flex>
      </Pad>
      <Pad wide center>
        <a target="_blank" href={codeRes.validationURI} rel="noreferrer">
          <Button
            type="button"
            startIcon={<Icon size="base" type={IconType.ExternalTab} />}
          >
            Authorize Github Access
          </Button>
        </a>
      </Pad>
      <Pad wide center>
        <div>Waiting for authorization to be completed...</div>
      </Pad>
    </div>
  );
})`
  .copy-code {
    border-radius: 4px;
    border: none;
    padding: 5px;
    min-width: auto;
    border-radius: 50%;
    &:hover {
      border: none;
    }
  }
  .code-text {
    font-size: ${extraLarge};
    margin: 0px 5px 0px 0px;
  }
`;
export default ModalContent;
