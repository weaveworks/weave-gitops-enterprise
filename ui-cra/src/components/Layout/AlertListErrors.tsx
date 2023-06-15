import { Box, Button, Collapse } from '@material-ui/core';
import Alert from '@material-ui/lab/Alert';
import { Flex, Icon, IconType, Text } from '@weaveworks/weave-gitops';
import { sortBy, uniqBy } from 'lodash';
import { FC, useEffect, useState } from 'react';
import styled from 'styled-components';
import { ReactComponent as ErrorIcon } from '../../assets/img/error.svg';
import { ListError } from '../../cluster-services/cluster_services.pb';

const BoxWrapper = styled(Box)`
  .MuiAlert-root {
    margin-bottom: ${props => props.theme.spacing.base};
    background: ${props => props.theme.colors.alertLight};
    border-radius: ${props => props.theme.spacing.xs};
  }
  .MuiAlert-action {
    display: inline;
    color: ${props => props.theme.colors.alertMedium};
  }
  .MuiIconButton-root:hover {
    background-color: ${props => props.theme.colors.alertLight};
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
const ErrorText = styled(Text)`
  margin-left: 8px;
`;
const NavButton = styled(Button)`
  padding: 0;
  min-width: auto;
  margin: 0;
`;
const ErrorsCount = styled.span`
  background: ${props => props.theme.colors.feedbackMedium};
  color: ${props => props.theme.colors.white};
  padding: 4px;
  border-radius: 4px;
  margin: 0 4px;
`;
export const AlertListErrors: FC<{ errors?: ListError[] }> = ({ errors }) => {
  const [index, setIndex] = useState<number>(0);
  const [filteredErrors, setFilteredErrors] = useState<ListError[]>([]);
  const [show, setShow] = useState<boolean>(true);

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
            <Flex align center>
              <ErrorIcon />
              <ErrorText size="base" data-testid="error-message">
                {filteredErrors[index].clusterName}:&nbsp;
                {filteredErrors[index].message}
              </ErrorText>
            </Flex>
            <Flex align center>
              <NavButton
                disabled={index === 0}
                data-testid="prevError"
                onClick={() => setIndex(currIndex => currIndex - 1)}
              >
                <Icon
                  type={IconType.NavigateBeforeIcon}
                  color="alertMedium"
                  size="medium"
                />
              </NavButton>
              <ErrorsCount data-testid="errorsCount">
                {filteredErrors.length}
              </ErrorsCount>
              <NavButton
                disabled={filteredErrors.length === index + 1}
                id="nextError"
                data-testid="nextError"
                onClick={() => setIndex(currIndex => currIndex + 1)}
              >
                <Icon
                  type={IconType.NavigateNextIcon}
                  color="alertMedium"
                  size="medium"
                />
              </NavButton>
            </Flex>
          </Alert>
        )}
      </Collapse>
    </BoxWrapper>
  );
};
