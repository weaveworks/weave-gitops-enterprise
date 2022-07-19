import React, { FC, Dispatch } from 'react';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import TextareaAutosize from '@material-ui/core/TextareaAutosize';
import Dialog from '@material-ui/core/Dialog';
import DialogTitle from '@material-ui/core/DialogTitle';
import { CloseIconButton } from '../../../../../assets/img/close-icon-button';
import DialogContent from '@material-ui/core/DialogContent';
import Typography from '@material-ui/core/Typography';

const xs = weaveTheme.spacing.xs;

const useStyles = makeStyles(() =>
  createStyles({
    textarea: {
      width: '100%',
      padding: xs,
      border: `1px solid ${weaveTheme.colors.neutral20}`,
    },
  }),
);

const Preview: FC<{
  openPreview: boolean;
  setOpenPreview: Dispatch<React.SetStateAction<boolean>>;
  PRPreview: string;
}> = ({ PRPreview, openPreview, setOpenPreview }) => {
  const classes = useStyles();

  return (
    <Dialog
      open={openPreview}
      maxWidth="md"
      fullWidth
      scroll="paper"
      onClose={() => setOpenPreview(false)}
    >
      <DialogTitle disableTypography>
        <Typography variant="h5">PR Preview</Typography>
        <CloseIconButton onClick={() => setOpenPreview(false)} />
      </DialogTitle>
      <DialogContent>
        <TextareaAutosize
          className={classes.textarea}
          value={PRPreview}
          readOnly
        />
        <span>
          You may edit these as part of the pull request with your git provider.
        </span>
      </DialogContent>
    </Dialog>
  );
};

export default Preview;
