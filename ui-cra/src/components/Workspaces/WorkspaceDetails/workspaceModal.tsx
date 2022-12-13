import { FC } from 'react';
import { Typography, DialogContent, DialogTitle } from '@material-ui/core';
import { CloseIconButton } from '../../../assets/img/close-icon-button';
import { WorkspaceRoleRule } from '../../../cluster-services/cluster_services.pb';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { DialogWrapper, RulesList } from '../WorkspaceStyles';

interface Props {
  onFinish: () => void;
  title: string;
  contentType: string;
  content: string | WorkspaceRoleRule[];
}
const WorkspaceModal: FC<Props> = ({
  onFinish,
  title,
  contentType,
  content,
}) => {
  const GetContent = () => {
    switch (contentType) {
      case 'yaml':
        return (
          <SyntaxHighlighter
            language="yaml"
            wrapLongLines="pre-wrap"
            showLineNumbers
          >
            {content}
          </SyntaxHighlighter>
        );
      case 'rules':
        return (
          <RulesList>
            {Array.isArray(content) &&
              content.map((rule: WorkspaceRoleRule, index: number) => (
                <li key={index}>
                  <div>
                    <label>Resources:</label>
                    <span>{rule.resources?.join(', ')}</span>
                  </div>
                  <div>
                    <label>Verbs:</label>
                    <span>{rule.verbs?.join(', ')}</span>
                  </div>
                  <div>
                    <label>Api Groups:</label>
                    <span>{rule.groups?.join('.')}</span>
                  </div>
                </li>
              ))}
          </RulesList>
        );
      default:
        return <span></span>;
    }
  };
  return (
    <DialogWrapper
      open
      maxWidth="md"
      fullWidth
      scroll="paper"
      onClose={() => onFinish()}
    >
      <DialogTitle disableTypography>
        <div>
          <Typography>{title}</Typography>
          <CloseIconButton onClick={() => onFinish()} />
        </div>
        {contentType === 'yaml' && (
          <span className="info">
            [some command related to retrieving this yaml]
          </span>
        )}
      </DialogTitle>
      <DialogContent
        className={contentType === 'rules' ? 'customBackgroundColor' : ''}
      >
        {GetContent()}
      </DialogContent>
    </DialogWrapper>
  );
};

export default WorkspaceModal;