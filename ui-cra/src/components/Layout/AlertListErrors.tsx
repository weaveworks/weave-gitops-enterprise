import { FC, useEffect, useState } from 'react';
import { ListError } from '../../cluster-services/cluster_services.pb';
import { theme } from '@weaveworks/weave-gitops';
import { Button, createStyles, makeStyles } from '@material-ui/core';
import ArrowBackIosOutlined from '@material-ui/icons/ArrowBackIosOutlined';
import ArrowForwardIosOutlined from '@material-ui/icons/ArrowForwardIosOutlined';
import { uniqBy, sortBy } from 'lodash';
import styled from 'styled-components';

import { ReactComponent as ErrorIcon } from '../../assets/img/error.svg';

const { base, medium, xs, xxs } = theme.spacing;
const {  neutral00 } = theme.colors;

const useAlertStyles = makeStyles(() =>
  createStyles({
    navigationBtn: {
      padding: 0,
      minWidth: 'auto',
      margin:0,
    },
    errosCount: {
      background: '#D58572',
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
    },
  }),
);

const AlertWrapper = styled.div`
  padding: ${base} ${medium};
  margin: 0 ${base} ${base} ${base};
  border-radius: 10px;
  background: #eecec7;
  display: flex;
  justify-content: space-between;
  align-items: center;
`;
const FlexCenter = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
`;

export const AlertListErrors: FC<{ errors?: ListError[] }> = ({ errors }) => {
  const [index, setIndex] = useState<number>(0);
  const [filteredErrors, setFilteredErrors] = useState<ListError[]>([]);

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
    <AlertWrapper id="alert-list-errors">
      {!!filteredErrors[index] && (
        <>
          <FlexCenter>
            <ErrorIcon className={classes.alertIcon} />
            <div className={classes.errorMessage} data-testid="error-message">
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
        </>
      )}
    </AlertWrapper>
  );
};
