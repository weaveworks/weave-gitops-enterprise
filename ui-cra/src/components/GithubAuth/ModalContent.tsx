import { useCallback, useState } from 'react';
import styled from 'styled-components';
import { Flex, Icon, IconType, Button, theme } from '@weaveworks/weave-gitops';
import { useGetGithubAuthStatus } from '../../contexts/GithubAuth';

const { extraLarge } = theme.fontSizes;
const { black, primary20 } = theme.colors;

const Pad = styled(Flex)`
  padding: 8px 0;
`;
const Text = styled.span`
  font-size: ${extraLarge};
  display: inline-flex;
`;
const PointerIcon = styled(Icon)`
  cursor: pointer;
`;

const CopyToClipboard = ({ value }: { value: string }) => {
  const [copied, setCopied] = useState(false);
  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(value).then(() => {
      setCopied(true);
      setTimeout(() => {
        setCopied(false);
      }, 3000);
    });
  }, [value]);

  return (
    <Text onClick={handleCopy}>
      {value}
      <PointerIcon
        type={copied ? IconType.CheckMark : IconType.FileCopyIcon}
        color={copied ? primary20 : black}
        size="medium"
      />
    </Text>
  );
};
const ModalContent = styled(({ codeRes, onSuccess, className }: any) => {
  const { isLoading, data, error } = useGetGithubAuthStatus(codeRes);
  if (!!data) {
    onSuccess(data.accessToken);
  }
  return (
    <div className={className}>
      <Pad wide center>
        <CopyToClipboard value={codeRes.userCode as string} />
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
        {isLoading && <div>Waiting for authorization to be completed...</div>}
        {error && (
          <div
            style={{
              color: 'red',
            }}
          >
            {error.message}
          </div>
        )}
      </Pad>
    </div>
  );
})`
  ${Icon} {
    margin-left: 8px;
  }
`;
export default ModalContent;
