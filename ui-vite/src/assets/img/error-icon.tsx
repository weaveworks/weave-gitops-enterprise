import React, { FC } from 'react';
import styled from 'styled-components';
import { fontSize } from 'weaveworks-ui-components/lib/theme/selectors';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faExclamationTriangle } from '@fortawesome/free-solid-svg-icons';

import { blinking } from '../effects/blinking';

const Icon = styled(FontAwesomeIcon)`
  ${blinking}
  color: ${({ theme }) => theme.colors.orange600};
  font-size: ${fontSize('normal')};
  display: inline-block;
  height: 16px;
`;

interface ErrorIconProps {
  error: string;
}

export const ErrorIcon: FC<ErrorIconProps> = ({ error }) => (
  <Icon icon={faExclamationTriangle} title={error} />
);
