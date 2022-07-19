import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import { FluxObjectRef } from '@weaveworks/weave-gitops/ui/lib/api/core/types.pb';
import * as React from 'react';
import styled from 'styled-components';

type Props = {
  className?: string;
  yaml: string;
  object: FluxObjectRef;
  kind: string;
};

export const IconButton = styled(Button)`
  &.MuiButton-root {
    border-radius: 50%;
    min-width: 48px;
    height: 48px;
    padding: 0;
  }
  &.MuiButton-text {
    padding: 0;
  }
`;

const YamlHeader = styled.div`
  background: ${props => props.theme.colors.neutral10};
  padding: ${props => props.theme.spacing.small};
  width: 100%;
  border-bottom: 1px solid ${props => props.theme.colors.neutral20};
  font-family: monospace;
  color: ${props => props.theme.colors.primary};
  text-overflow: ellipsis;
`;

const CopyButton = styled(IconButton)`
  &.MuiButton-outlinedPrimary {
    border: 1px solid ${props => props.theme.colors.neutral10};
    padding: ${props => props.theme.spacing.xs};
  }
  &.MuiButton-root {
    height: initial;
    width: initial;
    min-width: 0px;
  }
`;

function YamlView({ yaml, object, className, kind }: Props) {
  const [copied, setCopied] = React.useState(false);
  const headerText = `kubectl get ${kind?.toLowerCase()} ${object.name} -n ${
    object.namespace
  } -o yaml `;

  return (
    <div className={className}>
      <YamlHeader>
        {headerText}
        <CopyButton
          onClick={() => {
            navigator.clipboard.writeText(headerText);
            setCopied(true);
          }}
        >
          <Icon
            type={copied ? IconType.CheckMark : IconType.FileCopyIcon}
            size="small"
          />
        </CopyButton>
      </YamlHeader>
      <pre>
        {yaml.split('\n').map((yaml, index) => (
          <code key={index}>{yaml}</code>
        ))}
      </pre>
    </div>
  );
}

export default styled(YamlView).attrs({
  className: YamlView.name,
})`
  margin-bottom: ${props => props.theme.spacing.small};
  width: calc(100% - ${props => props.theme.spacing.medium});
  font-size: ${props => props.theme.fontSizes.small};
  border: 1px solid ${props => props.theme.colors.neutral20};
  border-radius: 8px;
  overflow: scroll;
  pre {
    padding: ${props => props.theme.spacing.small};
    white-space: pre-wrap;
  }

  pre::before {
    counter-reset: listing;
  }

  code {
    width: 100%;
    counter-increment: listing;
    text-align: left;
    float: left;
    clear: left;
  }

  code::before {
    width: 28px;
    color: ${props => props.theme.colors.primary};
    content: counter(listing);
    float: left;
    height: auto;
    padding-left: auto;
    margin-right: ${props => props.theme.spacing.small};
    text-align: right;
  }

  ${CopyButton} {
    .MuiButton-root {
      border-radius: 50%;
      min-width: 48px;
      height: 48px;
      padding: 0;
    }

    .MuiButton-text {
      padding: 0;
    }
  }
`;
