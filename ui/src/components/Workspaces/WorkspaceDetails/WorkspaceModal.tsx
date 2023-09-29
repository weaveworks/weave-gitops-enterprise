import { DialogContent } from '@material-ui/core';
import { Button, Text } from '@weaveworks/weave-gitops';
import { FC, useState } from 'react';
import styled from 'styled-components';
import { MuiDialogTitle } from '../../Shared';
import { DialogWrapper } from '../WorkspaceStyles';

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
  return (
    <>
      {title !== 'Rules' && (
        <Text color="primary" onClick={() => setIsModalOpen(true)} pointer>
          {btnName}
        </Text>
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
          <MuiDialogTitle
            title={title}
            onFinish={() => setIsModalOpen(false)}
          />
          {caption && <span className="info">{caption}</span>}
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
