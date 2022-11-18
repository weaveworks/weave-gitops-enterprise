import { Flex, Icon, IconType } from '@weaveworks/weave-gitops';
import React from 'react';
import styled from 'styled-components';
import { IconButton } from '../YamlView';
interface Props {
  command: string;
}

const CommandText = styled.p`
  margin: 0;
  margin-right: ${props => props.theme.spacing.base};
  white-space: pre;
`;

const CommandCell = ({ command }: Props) => {
  const [copied, setCopied] = React.useState(false);
  return (
    <Flex align>
      <CommandText>{command.replace('--', '\\\n --')}</CommandText>
      <IconButton
        onClick={() => {
          navigator.clipboard.writeText(command);
          setCopied(true);
          setTimeout(() => setCopied(false), 2000);
        }}
        variant="text"
      >
        <Icon
          type={copied ? IconType.CheckMark : IconType.FileCopyIcon}
          size="base"
          color="neutral40"
        />
      </IconButton>
    </Flex>
  );
};

export default CommandCell;
