import React, { FC } from 'react';
import styled from 'styled-components';
import { ClusterStatus } from '../../types/kubernetes';
import Box from '@material-ui/core/Box';
import moment from 'moment';
import { blinking } from '../../assets/effects/blinking';
import theme from 'weaveworks-ui-components/lib/theme';

const StrikeThrough = styled.line`
  ${blinking}
  stroke: ${theme.colors.white};
  stroke-width: 16px;
`;

const ReadyStatusIcon: FC<{
  color: HexColor;
  filledIn: boolean;
  strikeThrough?: boolean;
}> = ({ color, filledIn, strikeThrough }) => (
  <svg height="15px" style={{ marginRight: '2px' }} viewBox="0 0 100 100">
    <circle
      cx="50"
      cy="50"
      fill={filledIn ? color : 'transparent'}
      r="25"
      stroke={color}
      strokeWidth={17}
    />
    {strikeThrough ? <StrikeThrough x1="0" y1="50" x2="100" y2="50" /> : null}
  </svg>
);

const ReadyStatusWrapper = styled.div`
  display: flex;
  align-items: center;
`;

const SecondaryStatusWrapper = styled(ReadyStatusWrapper)`
  margin-left: 4px;
  opacity: 0.6;
`;

export enum Status {
  alerting = 'Alerting',
  critical = 'Critical alerts',
  notConnected = 'Not connected',
  ready = 'Ready',
  lastSeen = 'Last seen',
}

interface ReadyStatusProps {
  status: Status;
  updatedAt?: string;
  showConnectedStatus?: boolean;
}

const green = '#27AE60';

export const statusSummary = (status: Status, updatedAt?: string): string =>
  updatedAt ? `Last seen ${moment.utc(updatedAt).format()}` : '';

export const ReadyStatus: FC<ReadyStatusProps> = ({
  status,
  updatedAt,
  showConnectedStatus,
}) => {
  const color: HexColor =
    (
      {
        'Not connected': '#BDBDBD',
        Alerting: '#F2994A',
        'Critical alerts': '#BC3B1D',
        Ready: green,
        'Last seen': '#BDBDBD',
      } as { [status in Status]: HexColor }
    )[status] || '#BDBDBD';

  const filledIn: boolean =
    (
      {
        'Not connected': true,
        Alerting: true,
        'Critical alerts': true,
        Ready: true,
        'Last seen': true,
      } as { [status in Status]: boolean }
    )[status] || false;

  const strikeThrough: boolean =
    (
      {
        'Not connected': false,
        Alerting: false,
        'Critical alerts': false,
        Ready: false,
        'Last seen': true,
      } as { [status in Status]: boolean }
    )[status] || false;

  const ConnectionStatusWrapper: FC = ({ children }) =>
    showConnectedStatus &&
    status !== Status.notConnected &&
    status !== Status.lastSeen ? (
      <>
        <ReadyStatusIcon color={green} filledIn />
        Connected <SecondaryStatusWrapper>({children})</SecondaryStatusWrapper>
      </>
    ) : (
      // https://github.com/DefinitelyTyped/DefinitelyTyped/issues/44572
      (children as unknown as JSX.Element)
    );

  return (
    <ReadyStatusWrapper>
      <ConnectionStatusWrapper>
        <ReadyStatusIcon
          color={color}
          filledIn={filledIn}
          strikeThrough={strikeThrough}
        />
        {status}{' '}
        {updatedAt && status === 'Last seen' && (
          <Box title={updatedAt} ml={1} color="text.secondary">
            {moment(updatedAt).fromNow()}
          </Box>
        )}
      </ConnectionStatusWrapper>
    </ReadyStatusWrapper>
  );
};

export const getClusterStatus = (status?: ClusterStatus) => {
  switch (status) {
    case 'notConnected':
      return Status.notConnected;

    case 'critical':
      return Status.critical;

    case 'ready':
      return Status.ready;

    case 'lastSeen':
      return Status.lastSeen;

    case 'alerting':
      return Status.alerting;

    default:
      return Status.notConnected;
  }
};
