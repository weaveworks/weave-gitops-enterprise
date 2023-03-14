import {
  Box,
  Button,
  Collapse,
  createStyles,
  makeStyles,
} from '@material-ui/core';
import {
  ArrowBackIosOutlined,
  ArrowForwardIosOutlined,
} from '@material-ui/icons';
import Alert from '@material-ui/lab/Alert';
import { theme } from '@weaveworks/weave-gitops';
import { sortBy, uniqBy } from 'lodash';
import { FC, useEffect, useState } from 'react';
import styled from 'styled-components';
import ErrorIcon from '../../assets/img/error.svg';
import { ListError } from '../../cluster-services/cluster_services.pb';

const { base, xs, xxs } = theme.spacing;
const { neutral00, alertLight, alertMedium } = theme.colors;

const useAlertStyles = makeStyles(() =>
  createStyles({
    navigationBtn: {
      padding: 0,
      minWidth: 'auto',
      margin: 0,
    },
    errosCount: {
      background: alertMedium,
      color: neutral00,
      padding: xxs,
      borderRadius: xxs,
      margin: `0 ${xxs}`,
    },
    alertIcon: {
      marginRight: xs,
    },
    errorMessage: {
      fontSize: base,
    },
    arrowIcon: {
      fontSize: '18px',
      fontWeight: 400,
      color: alertMedium,
    },
  }),
);

const BoxWrapper = styled(Box)`
  .MuiAlert-root {
    margin-bottom: ${base};
    background: #eecec7;
    border-radius: ${xs};
  }
  .MuiAlert-action {
    display: inline;
    color: ${alertMedium};
  }
  .MuiIconButton-root:hover {
    background-color: ${alertLight};
  }
  .MuiAlert-icon {
    .MuiSvgIcon-root {
      display: none;
    }
  }
  .MuiAlert-message {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
`;

const FlexCenter = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
`;

export const AlertListErrors: FC<{ errors?: ListError[] }> = ({ errors }) => {
  const [index, setIndex] = useState<number>(0);
  const [filteredErrors, setFilteredErrors] = useState<ListError[]>([]);
  const [show, setShow] = useState<boolean>(true);

  const classes = useAlertStyles();

  useEffect(() => {
    const fErrors = sortBy(
      uniqBy(errors, error => [error.clusterName, error.message].join()),
      [v => v.clusterName, v => v.namespace, v => v.message],
    );
    setFilteredErrors(fErrors);
    setIndex(0);
    return () => {
      setFilteredErrors([]);
    };
  }, [errors]);

  if (!errors || !errors.length) {
    return null;
  }

  return (
    <BoxWrapper id="alert-list-errors">
      <Collapse in={show}>
        {!!filteredErrors[index] && (
          <Alert severity="error" onClose={() => setShow(false)}>
            <FlexCenter>
              <ErrorIcon className={classes.alertIcon} />
              <div className={classes.errorMessage} data-testid="error-message">
                {filteredErrors[index].clusterName}:&nbsp;
                {filteredErrors[index].message}
              </div>
            </FlexCenter>
            <FlexCenter>
              <Button
                disabled={index === 0}
                className={classes.navigationBtn}
                data-testid="prevError"
                onClick={() => setIndex(currIndex => currIndex - 1)}
              >
                <ArrowBackIosOutlined className={classes.arrowIcon} />
              </Button>
              <span className={classes.errosCount} data-testid="errorsCount">
                {filteredErrors.length}
              </span>
              <Button
                disabled={filteredErrors.length === index + 1}
                className={classes.navigationBtn}
                id="nextError"
                data-testid="nextError"
                onClick={() => setIndex(currIndex => currIndex + 1)}
              >
                <ArrowForwardIosOutlined className={classes.arrowIcon} />
              </Button>
            </FlexCenter>
          </Alert>
        )}
      </Collapse>
    </BoxWrapper>
  );
};
