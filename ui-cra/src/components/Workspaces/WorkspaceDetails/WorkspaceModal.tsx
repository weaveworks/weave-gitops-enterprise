import { FC, useState } from 'react';
import { Typography, DialogContent, DialogTitle } from '@material-ui/core';
import { CloseIconButton } from '../../../assets/img/close-icon-button';
import {
  DialogWrapper,
  useWorkspaceStyle,
  ViewYamlBtn,
} from '../WorkspaceStyles';
import { Button } from '@weaveworks/weave-gitops';

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
      {content && (
        <div className={title !== 'Rules' ? classes.YamlBtn : ''}>
          <Button
            style={{ marginRight: 0, textTransform: 'uppercase' }}
            onClick={() => setIsModalOpen(true)}
          >
            {btnName}
          </Button>
        </div>
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
