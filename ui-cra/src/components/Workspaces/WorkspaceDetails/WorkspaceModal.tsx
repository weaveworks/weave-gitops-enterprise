import { DialogContent, DialogTitle, Typography } from '@material-ui/core';
import { Button } from '@weaveworks/weave-gitops';
import { FC, useState } from 'react';
import { CloseIconButton } from '../../../assets/img/close-icon-button';
import { DialogWrapper, useWorkspaceStyle } from '../WorkspaceStyles';

interface Props {
  title: string;
  caption?: string;
  content: any;
  className?: string;
  btnName: string;
}
const WorkspaceModal: FC<Props> = ({
  title,
  caption,
  content,
  className,
  btnName,
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
          <DialogContent className={className || ''}>{content}</DialogContent>
        </DialogWrapper>
      )}
    </>
  );
};

export default WorkspaceModal;
