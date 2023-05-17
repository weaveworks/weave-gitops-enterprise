import { FC, useState } from 'react';
import { Typography, DialogContent, DialogTitle } from '@material-ui/core';
import styled from 'styled-components';
import { CloseIconButton } from '../../../assets/img/close-icon-button';
import { DialogWrapper, useWorkspaceStyle } from '../WorkspaceStyles';
import { Button } from '@weaveworks/weave-gitops';

interface Props {
  title: string;
  caption?: string;
  content: any;
  className?: string;
  btnName: string;
  wrapDialogContent?: boolean;
}

const ContentWrapper = styled.div`
  padding: ${({ theme }) => theme.spacing.medium}
    ${({ theme }) => theme.spacing.small};
  overflow-y: auto;
`;

const WorkspaceModal: FC<Props> = ({
  title,
  caption,
  content,
  className,
  btnName,
  wrapDialogContent,
}) => {
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);
  const classes = useWorkspaceStyle();

  return (
    <>
      {title !== 'Rules' && (
        <span className={classes.link} onClick={() => setIsModalOpen(true)}>
          {btnName}
        </span>
      )}
      {title === 'Rules' && content && (
        <Button
          style={{ marginRight: 0, textTransform: 'uppercase' }}
          onClick={() => setIsModalOpen(true)}
        >
          {btnName}
        </Button>
      )}

      {isModalOpen && (
        <DialogWrapper
          open={isModalOpen}
          maxWidth="md"
          fullWidth
          scroll="paper"
        >
          <DialogTitle disableTypography>
            <div>
              <Typography>{title}</Typography>
              <CloseIconButton onClick={() => setIsModalOpen(false)} />
            </div>
            {caption && <span className="info">{caption}</span>}
          </DialogTitle>
          {wrapDialogContent ? (
            <DialogContent className={className || ''}>{content}</DialogContent>
          ) : (
            <ContentWrapper className={className || ''}>
              {content}
            </ContentWrapper>
          )}
        </DialogWrapper>
      )}
    </>
  );
};

export default WorkspaceModal;
