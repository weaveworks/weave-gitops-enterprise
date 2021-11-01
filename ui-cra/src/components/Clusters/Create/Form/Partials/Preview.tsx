import React, { FC, Dispatch } from 'react';
import { FormStep } from '../Steps';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import TextareaAutosize from '@material-ui/core/TextareaAutosize';

const xs = weaveTheme.spacing.xs;

const useStyles = makeStyles(() =>
  createStyles({
    textarea: {
      width: '100%',
      padding: xs,
      border: '1px solid #E5E5E5',
    },
  }),
);

const Preview: FC<{
  activeStep: string | undefined;
  setActiveStep: Dispatch<React.SetStateAction<string | undefined>>;
  clickedStep: string;
  PRPreview: string;
}> = ({ activeStep, setActiveStep, clickedStep, PRPreview }) => {
  const classes = useStyles();

  return (
    <FormStep
      title="Preview"
      active={activeStep === 'Preview'}
      clicked={clickedStep === 'Preview'}
      setActiveStep={setActiveStep}
    >
      <TextareaAutosize
        className={classes.textarea}
        value={PRPreview}
        readOnly
      />
      <span>
        You may edit these as part of the pull request with your git provider.
      </span>
    </FormStep>
  );
};

export default Preview;
