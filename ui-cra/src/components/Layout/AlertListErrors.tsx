import { FC, useEffect, useState } from 'react';
import { ListError } from '../../cluster-services/cluster_services.pb';
import { theme } from '@weaveworks/weave-gitops';
import {
  Button,
  createStyles,
  makeStyles,
  IconButton,
  Box,
  Collapse,
} from '@material-ui/core';
import Alert from '@material-ui/lab/Alert';
import {
  ArrowBackIosOutlined,
  ArrowForwardIosOutlined,
  Close,
} from '@material-ui/icons';
import { uniqBy, sortBy } from 'lodash';
import styled from 'styled-components';

import { ReactComponent as ErrorIcon } from '../../assets/img/error.svg';

const { base, medium, xs, xxs } = theme.spacing;
const { neutral00 } = theme.colors;

const useAlertStyles = makeStyles(() =>
  createStyles({
    navigationBtn: {
      padding: 0,
      minWidth: 'auto',
      margin: 0,
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

const AlertWrapper = styled(Alert)`
  .MuiAlert-action {
    display: inline;
  }
  // display: flex;
  // justify-content: space-between;
  // align-items: center;
  padding: ${base} ${medium};
  margin: 0 ${base} ${base} ${base};
  border-radius: 10px;
  background: #eecec7;
  .MuiIconButton-root:hover {
    background-color: #eecec7;
  }
`;

const AlertList = styled.div`
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
    <Box id="alert-list-errors">
      <Collapse in={show}>
        {!!filteredErrors[index] && (
          <AlertWrapper severity="error" onClose={() => setShow(false)}>
            <AlertList>
              <FlexCenter>
                <ErrorIcon className={classes.alertIcon} />
                <div
                  className={classes.errorMessage}
                  data-testid="error-message"
                >
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
            </AlertList>
          </AlertWrapper>
        )}
      </Collapse>
    </Box>
  );
  // show ? (
  // <AlertWrapper id="alert-list-errors">
  //   {!!filteredErrors[index] && (
  //       <>
  //         <FlexCenter>
  //           <ErrorIcon className={classes.alertIcon} />
  //           <div className={classes.errorMessage} data-testid="error-message">
  //             {filteredErrors[index].clusterName}:&nbsp;
  //             {filteredErrors[index].message}
  //           </div>
  //         </FlexCenter>

  //         <FlexCenter>
  //           <Button
  //             disabled={index === 0}
  //             className={classes.navigationBtn}
  //             data-testid="prevError"
  //             onClick={() => setIndex(currIndex => currIndex - 1)}
  //           >
  //             <ArrowBackIosOutlined className={classes.arrowIcon} />
  //           </Button>
  //           <span className={classes.errosCount} data-testid="errorsCount">
  //             {filteredErrors.length}
  //           </span>
  //           <Button
  //             disabled={filteredErrors.length === index + 1}
  //             className={classes.navigationBtn}
  //             id="nextError"
  //             data-testid="nextError"
  //             onClick={() => setIndex(currIndex => currIndex + 1)}
  //           >
  //             <ArrowForwardIosOutlined className={classes.arrowIcon} />
  //           </Button>
  //         </FlexCenter>
  //       </>
  //   )}
  //   <IconButton
  //     style={{
  //       paddingTop: 0,
  //       paddingRight: 0,
  //     }}
  //     onClick={() => setShow(false)}
  //   >
  //     <Close />
  //   </IconButton>
  // </AlertWrapper>

  // ) : null;
};
