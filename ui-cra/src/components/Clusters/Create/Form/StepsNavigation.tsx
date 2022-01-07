import React, { Dispatch, FC, useCallback } from 'react';
import Link from '@material-ui/core/Link';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import classNames from 'classnames';

const useStyles = makeStyles(() =>
  createStyles({
    navWrapper: {
      position: 'sticky',
      top: '112px',
      overflow: 'auto',
    },
    linkWrapper: {
      display: 'flex',
      justifyContent: 'flex-end',
      whiteSpace: 'nowrap',
    },
    link: {
      fontSize: `${weaveTheme.fontSizes.tiny}`,
      padding: `0 ${weaveTheme.spacing.small}`,
      marginBottom: `${weaveTheme.spacing.small}`,
      color: `${weaveTheme.colors.black}`,
      cursor: 'pointer',
    },
    activeLink: {
      borderRight: `4px solid ${weaveTheme.colors.primary}`,
    },
  }),
);

const FormStepsNavigation: FC<{
  steps: string[];
  activeStep: string | undefined;
  setClickedStep: Dispatch<React.SetStateAction<string>>;
  PRPreview?: string | null;
}> = ({ steps, activeStep, setClickedStep, PRPreview }) => {
  const classes = useStyles();

  const handleClick = useCallback(
    (event: React.MouseEvent<HTMLAnchorElement, MouseEvent>) => {
      event.preventDefault();
      setClickedStep(event.currentTarget.text);
    },
    [setClickedStep],
  );

  const sections = PRPreview
    ? [...steps, 'Profiles', 'Preview', 'GitOps']
    : steps && [...steps, 'Profiles'];

  return (
    <div className={classes.navWrapper}>
      {sections?.map((step, index) => {
        const kindStep = step.split(' ')[0];
        const displayNameStep = step.replace(kindStep, '');

        return (
          <div key={index} className={classes.linkWrapper}>
            <Link
              style={{
                textDecoration: 'none',
              }}
              className={classNames(
                classes.link,
                step === activeStep ? classes.activeLink : '',
              )}
              onClick={handleClick}
            >
              <div>{kindStep}</div>
              <div>{displayNameStep}</div>
            </Link>
          </div>
        );
      })}
    </div>
  );
};

export default FormStepsNavigation;
