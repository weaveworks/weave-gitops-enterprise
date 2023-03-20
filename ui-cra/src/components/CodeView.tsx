import {
  Canary,
  CanaryMetricTemplate,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { CopyToClipboard, createYamlCommand } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

type Props = {
  className?: string;
  code: string;
  object?: Canary | CanaryMetricTemplate;
  kind?: string;
  colorizeChanges?: boolean;
};

const YamlHeader = styled.div`
  background: ${props => props.theme.colors.neutral10};
  padding: ${props => props.theme.spacing.small};
  width: 100%;
  border-bottom: 1px solid ${props => props.theme.colors.neutral20};
  font-family: monospace;
  color: ${props => props.theme.colors.primary};
  text-overflow: ellipsis;
`;

const additionColor = '#2e5f38';
const deletionColor = '#814a1c';
const updateColor = '#787118';

function CodeView({ code, object, className, kind, colorizeChanges }: Props) {
  let headerText = createYamlCommand(
    kind || '',
    object?.name || '',
    object?.namespace || '',
  );

  return (
    <div className={className}>
      {headerText && (
        <YamlHeader>
          {headerText}
          <CopyToClipboard size="small" value={headerText}></CopyToClipboard>
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
  height: 100%;
  overflow: auto;
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
