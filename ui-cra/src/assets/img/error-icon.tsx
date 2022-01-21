import React, { FC } from 'react';
import styled from 'styled-components';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faExclamationTriangle } from '@fortawesome/free-solid-svg-icons';
import { blinking } from '../effects/blinking';
import { theme } from '@weaveworks/weave-gitops';

const Icon = styled(FontAwesomeIcon)`
  ${blinking}
  color: ${({ theme }) => theme.colors.alert};
  font-size: ${theme.fontSizes.normal};
  display: inline-block;
  height: 16px;
`;

interface ErrorIconProps {
  error: string;
}

export const ErrorIcon: FC<ErrorIconProps> = ({ error }) => (
  <Icon icon={faExclamationTriangle} title={error} />
);
