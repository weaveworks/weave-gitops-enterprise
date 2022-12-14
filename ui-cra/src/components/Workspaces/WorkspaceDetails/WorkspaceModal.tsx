import { FC, useState } from 'react';
import { Typography, DialogContent, DialogTitle } from '@material-ui/core';
import { CloseIconButton } from '../../../assets/img/close-icon-button';
import { DialogWrapper, ViewYamlBtn } from '../WorkspaceStyles';
import { Button } from '@weaveworks/weave-gitops';

interface Props {
  title: string;
  caption?: string;
  content: any;
  className?: string;
  btnName:string
}
const WorkspaceModal: FC<Props> = ({ title, caption, content, className, btnName }) => {
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);

  return (
    <>
      {content && (
        <ViewYamlBtn>
          <Button
            style={{ marginRight: 0, textTransform: 'uppercase' }}
            onClick={() => setIsModalOpen(true)}
          >
            {btnName}
          </Button>
        </ViewYamlBtn>
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
          <DialogContent className={className || ''}>
            {content}
          </DialogContent>
        </DialogWrapper>
      )}
    </>
  );
};

export default WorkspaceModal;
