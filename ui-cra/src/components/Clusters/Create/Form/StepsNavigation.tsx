import React, { Dispatch, FC } from 'react';
import Link from '@material-ui/core/Link';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import classNames from 'classnames';

const useStyles = makeStyles(() =>
  createStyles({
    navWrapper: {
      position: 'sticky',
      top: '112px',
    },
    linkWrapper: {
      display: 'flex',
      justifyContent: 'flex-end',
      flexWrap: 'wrap',
    },
    link: {
      fontSize: `${weaveTheme.fontSizes.small}`,
      padding: `0 ${weaveTheme.spacing.small}`,
      marginBottom: `${weaveTheme.spacing.small}`,
      color: `${weaveTheme.colors.black}`,
      cursor: 'pointer',
    },
    activeLink: {
      borderRight: `4px solid ${weaveTheme.colors.blue400}`,
    },
  }),
);

const FormStepsNavigation: FC<{
  steps: string[];
  activeStep: string;
  setActiveStep: Dispatch<React.SetStateAction<string>>;
}> = ({ steps, activeStep, setActiveStep }) => {
  const classes = useStyles();

  const handleClick = (event: any) => {
    event.preventDefault();
    setActiveStep(event.target.text);
  };

  return (
    <div className={classes.navWrapper}>
      {steps.map((step, index) => {
        return (
          <div key={index} className={classes.linkWrapper}>
            <Link
              style={{ textDecoration: 'none' }}
              className={classNames(
                classes.link,
                step === activeStep ? classes.activeLink : '',
              )}
              onClick={handleClick}
            >
              {step}
            </Link>
          </div>
        );
      })}
    </div>
  );
};

export default FormStepsNavigation;
