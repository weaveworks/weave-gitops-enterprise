import {
  Canary,
  CanaryMetricTemplate,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';

type Props = {
  className?: string;
  code: string;
  object?: Canary | CanaryMetricTemplate;
  kind?: string;
  colorizeChanges?: boolean;
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

const additionColor = '#2e5f38';
const deletionColor = '#814a1c';
const updateColor = '#787118';

function CodeView({ code, object, className, kind, colorizeChanges }: Props) {
  const [copied, setCopied] = React.useState(false);

  let headerText = '';

  if (kind && object) {
    headerText = `kubectl get ${kind.toLowerCase()} ${object.name} -n ${
      object.namespace
    } -o yaml `;
  }

  return (
    <div className={className}>
      {headerText && (
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
      )}

      <pre>
        {colorizeChanges
          ? code.split('\n').map((code, index) => {
              let color = '';

              if (/^\s+\+ /.test(code)) {
                color = additionColor;
              } else if (/^\s+- /.test(code)) {
                color = deletionColor;
              } else if (/^\s+~ /.test(code)) {
                color = updateColor;
              }

              return (
                <code key={index}>
                  {color ? (
                    <span
                      style={{
                        color: color,
                      }}
                    >
                      {code}
                    </span>
                  ) : (
                    code
                  )}
                </code>
              );
            })
          : code
              .split('\n')
              .map((code, index) => <code key={index}>{code}</code>)}
      </pre>
    </div>
  );
}

export default styled(CodeView).attrs({
  className: CodeView.name,
})`
  margin-bottom: ${props => props.theme.spacing.small};
  width: calc(100% - ${props => props.theme.spacing.medium});
  font-size: ${props => props.theme.fontSizes.small};
  border: 1px solid ${props => props.theme.colors.neutral20};
  border-radius: 8px;
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

  button {
    border: 0px;
    border-radius: 50%;
    min-width: 30px;
    height: 30px;
    padding: 0;

    &:hover {
      border: 1px solid #d8d8d8;
    }
  }
`;
