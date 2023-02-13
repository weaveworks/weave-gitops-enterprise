import {
  IconButton,
  MenuItem,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableRow,
} from '@material-ui/core';
import { Error, Info } from '@material-ui/icons';
import {
  Flex,
  formatLogTimestamp,
  Icon,
  IconType,
} from '@weaveworks/weave-gitops';
import { LogEntry } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import React, { SetStateAction } from 'react';
import styled from 'styled-components';
import { useGetLogs } from '../../../hooks/gitopsrun';
import { Select } from '../../../utils/form';

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

const makeHeader = (logs: LogEntry[], refetching: boolean) => {
  if (!logs.length) {
    if (refetching) return 'Fetching logs...';
    else return 'No logs found';
  }

  const beginning = formatLogTimestamp(logs[0].timestamp);
  const end = formatLogTimestamp(logs[logs.length - 1].timestamp);

  return `showing logs from ${beginning} to ${end}`;
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
      <TableCell className="gray">
        {formatLogTimestamp(log.timestamp)}
      </TableCell>
      <TableCell>{log.source || '-'}</TableCell>
      <TableCell className="break-word">{log.message || '-'}</TableCell>
    </TableRow>
  );
};

function GitOpsRunLogs({ className, name, namespace }: Props) {
  const [reverseSort, setReverseSort] = React.useState<boolean>(false);
  const [token, setToken] = React.useState<string>('');
  const [logValue, setLogValue] = React.useState<string>('all');
  const [levelValue, setLevelValue] = React.useState<string>('all');
  const [logSources, setLogSources] = React.useState<string[]>([]);
  const [logs, setLogs] = React.useState<LogEntry[]>([]);
  const [refetching, setRefetching] = React.useState<boolean>(false);
  const { isLoading, data, refetch } = useGetLogs({
    sessionNamespace: namespace,
    sessionId: name,
    token,
    logSourceFilter: logValue === 'all' ? '' : logValue,
    logLevelFilter: levelValue === 'all' ? '' : levelValue,
  });

  const refetchOnChange = (
    value: string,
    stateFunction: React.Dispatch<SetStateAction<string>>,
  ) => {
    const stateActions = new Promise(() => {
      setRefetching(true);
      //select dropdown value
      stateFunction(value);
      //reset logs request
      setLogs([]);
      setToken('' as string);
    });

    stateActions.then(() => {
      refetch();
      setRefetching(false);
    });
  };

  React.useEffect(() => {
    if (isLoading) return;
    if (data?.logs?.length && data?.nextToken) {
      const newLogs = data.logs;
      //if there are already logs in state
      if (logs.length)
        setLogs(
          reverseSort ? [...newLogs.reverse(), ...logs] : [...logs, ...newLogs],
        );
      else setLogs(reverseSort ? newLogs.reverse() : newLogs);
      setToken(data.nextToken);
      setLogSources(data?.logSources || []);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, data]);

  return (
    <Flex className={className} wide tall column>
      <Flex>
        <Select
          label="LOG"
          value={logValue}
          defaultValue={''}
          onChange={e => refetchOnChange(e.target.value as string, setLogValue)}
          className="pad-right"
        >
          <MenuItem key="all" value={'all'}>
            all
          </MenuItem>
          {logSources.map((source, index) => (
            <MenuItem key={index} value={source}>
              {source}
            </MenuItem>
          ))}
        </Select>
        <Select
          label="LEVEL"
          value={levelValue}
          defaultValue={'all'}
          onChange={e =>
            refetchOnChange(e.target.value as string, setLevelValue)
          }
        >
          <MenuItem key="all" value="all">
            all
          </MenuItem>
          <MenuItem key="info" value="info">
            info
          </MenuItem>
          <MenuItem key="warn" value="warn">
            warn
          </MenuItem>
          <MenuItem key="error" value="error">
            error
          </MenuItem>
        </Select>
      </Flex>
      <Header wide align>
        {makeHeader(logs, refetching)}
        <IconButton
          onClick={() => {
            setLogs(logs.reverse());
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
          <TableBody>
            {logs.map((log, index) => (
              <LogRow key={index} log={log} />
            ))}
          </TableBody>
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
  //adds padding left for Select text
  .MuiInputBase-input {
    padding: 6px 6px 7px;
  }
`;
