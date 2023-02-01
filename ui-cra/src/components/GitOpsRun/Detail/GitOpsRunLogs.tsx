import {
  IconButton,
  Table,
  TableCell,
  TableContainer,
  TableRow,
} from '@material-ui/core';
import { Error, Info } from '@material-ui/icons';
import { Flex, Icon, IconType } from '@weaveworks/weave-gitops';
import { LogEntry } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import React from 'react';
import styled from 'styled-components';
import { useGetLogs } from '../../../hooks/gitopsrun';

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

const Header = styled(Flex)`
  background: ${props => props.theme.colors.neutral10};
  padding: ${props => props.theme.spacing.xs};
  width: 100%;
  border-bottom: 1px solid ${props => props.theme.colors.neutral20};
  margin-bottom: ${props => props.theme.spacing.xxs};
`;

const makeHeader = (logs: LogEntry[], reverseSort: boolean) => {
  if (!logs.length) return 'No logs found';
  return `showing logs from ${
    reverseSort
      ? `now to ${logs[logs.length - 1].timestamp}`
      : `${logs[0].timestamp} to now`
  }`;
};

const LogRow: React.FC<{ log: LogEntry }> = ({ log }) => {
  return (
    <TableRow>
      <TableCell>
        <Flex /*this flex centers the icon*/>
          {log.level === 'info' ? (
            <Info color="primary" fontSize="inherit" />
          ) : (
            <Error color="secondary" fontSize="inherit" />
          )}
        </Flex>
      </TableCell>
      <TableCell className="gray">{log.timestamp || '-'}</TableCell>
      <TableCell>{log.source || '-'}</TableCell>
      <TableCell className="break-word">{log.message || '-'}</TableCell>
    </TableRow>
  );
};

function GitOpsRunLogs({ className, name, namespace }: Props) {
  // const [logOptions, setLogOptions] = React.useState<string[]>([
  //   'log one',
  //   'log two',
  // ]);
  // const [levelOptions, setLevelOptions] = React.useState<string[]>([
  //   'level one',
  //   'level two',
  // ]);
  // const [logValue, setLogValue] = React.useState('-');
  // const [levelValue, setLevelValue] = React.useState('-');

  const [reverseSort, setReverseSort] = React.useState(false);
  const [token, setToken] = React.useState('');
  const { isLoading, data, error } = useGetLogs({
    sessionNamespace: namespace,
    sessionId: name,
    token,
  });

  React.useEffect(() => {
    if (isLoading) return;
    setToken(data?.nextToken || '');
  }, [data]);

  const logs = data?.logs || [];

  return (
    <Flex className={className} wide tall column>
      {/* <Flex>
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
      </Flex> */}
      <Header wide align>
        {makeHeader(logs, reverseSort)}
        <IconButton
          onClick={() => {
            logs.reverse();
            setReverseSort(!reverseSort);
          }}
        >
          <Icon
            type={IconType.ArrowUpwardIcon}
            size="small"
            className={reverseSort ? 'upward' : 'downward'}
          />
        </IconButton>
      </Header>
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
  .MuiTableCell-root {
    line-height: 1;
    padding: 4px;
    border-bottom: none;
    white-space: nowrap;
    &.break-word {
      white-space: normal;
      word-break: break-word;
    }
    &.gray {
      color: ${props => props.theme.colors.neutral30};
    }
  }
  .MuiTableRow-root {
    border-bottom: none;
  }
`;
