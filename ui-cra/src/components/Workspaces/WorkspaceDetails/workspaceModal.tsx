import { FC } from 'react';
import {
  Typography,
  DialogContent,
  DialogTitle,
  Dialog,
} from '@material-ui/core';
import styled from 'styled-components';
import { CloseIconButton } from '../../../assets/img/close-icon-button';
import { WorkspaceRoleRule } from '../../../cluster-services/cluster_services.pb';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';

const DialogWrapper = styled(Dialog)`
  .MuiDialog-paper {
    border-top-right-radius: 10px;
    border-top-left-radius: 10px;
  }
  .MuiDialogTitle-root {
    background: ${({ theme }) => theme.colors.neutralGray};
    padding-top: ${({ theme }) => theme.spacing.medium};
    padding-left: ${({ theme }) => theme.spacing.medium};
    padding-bottom: ${({ theme }) => theme.spacing.small};
    p{
        font-weight: 600;
    }
    .MuiSvgIcon-root{
        color: ${({ theme }) => theme.colors.neutral30};

    }
    .info{
        color: ${({ theme }) => theme.colors.primary10} ;
        font-size: ${({ theme }) => theme.fontSizes.small};
        font-weight: 500;
    }
  }
  .MuiDialogContent-root{
    pre{
        background: #fff !important;
        padding-left:${({ theme }) => theme.spacing.none} !important;
        span{
        font-family: ${({ theme }) => theme.fontFamilies.monospace};
        font-size: ${({ theme }) => theme.fontSizes.small};
        text-align: left !important;
        padding-right: ${({ theme }) => theme.spacing.none} !important;
        min-width: 27px !important;
}
        }
    }
  }

`;

interface Props {
  onFinish: () => void;
  title: string;
  contentType: string;
  content: string | WorkspaceRoleRule[]
}
const WorkspaceModal: FC<Props> = ({
  onFinish,
  title,
  contentType,
  content
}) => {
  function GetContent() {
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
        return <ul>to be implemented</ul>;
      default:
        return <span>-</span>;
    }
  }
  console.log(content)
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
      <DialogContent>{GetContent()}</DialogContent>
    </DialogWrapper>
  );
};

export default WorkspaceModal;
