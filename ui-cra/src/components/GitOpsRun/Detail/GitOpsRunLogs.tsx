import { Table, TableCell, TableContainer, TableRow } from '@material-ui/core';
import { Error, Info } from '@material-ui/icons';
import { Flex } from '@weaveworks/weave-gitops';
import React from 'react';
import styled from 'styled-components';
import { Select } from '../../../utils/form';

type Log = {
  source: string;
  message: string;
  timestamp: string;
  severity: string;
};

type Props = {
  className?: string;
  logs: Log[];
};

const Header = styled(Flex)`
  background: ${props => props.theme.colors.neutral10};
  padding: ${props => props.theme.spacing.small};
  width: 100%;
  border-bottom: 1px solid ${props => props.theme.colors.neutral20};
`;

const LogRow: React.FC<{ log: Log }> = ({ log }) => {
  return (
    <TableRow>
      <TableCell>
        {log.severity === 'info' ? (
          <Info color="primary" />
        ) : (
          <Error color="secondary" />
        )}
      </TableCell>
      <TableCell>{log.timestamp || '-'}</TableCell>
      <TableCell>{log.source || '-'}</TableCell>
      <TableCell>{log.message || '-'}</TableCell>
    </TableRow>
  );
};

function GitOpsRunLogs({ className, logs }: Props) {
  const [logOptions] = React.useState<string[]>([
    'log one',
    'log two',
  ]);
  const [levelOptions] = React.useState<string[]>([
    'level one',
    'level two',
  ]);
  const [logValue, setLogValue] = React.useState('-');
  const [levelValue, setLevelValue] = React.useState('-');

  React.useEffect(() => {
    //find logs and levels for selects, plus earliest timestamp?!
  });

  return (
    <Flex className={className} wide tall column>
      <Flex>
        <Select
          label="LOG"
          value={logValue}
          items={logOptions}
          onChange={e => setLogValue(e.target.value as string)}
          className="pad-right"
        />
        <Select
          label="LEVEL"
          value={levelValue}
          items={levelOptions}
          onChange={e => setLevelValue(e.target.value as string)}
        />
      </Flex>
      <Header wide>showing logs from ....</Header>
      <TableContainer>
        <Table>
          {logs.map(log => (
            <LogRow log={log} />
          ))}
        </Table>
      </TableContainer>
    </Flex>
  );
}

export default styled(GitOpsRunLogs).attrs({ className: GitOpsRunLogs.name })`
  .pad-right {
    margin-right: ${props => props.theme.spacing.xs};
  }
`;
