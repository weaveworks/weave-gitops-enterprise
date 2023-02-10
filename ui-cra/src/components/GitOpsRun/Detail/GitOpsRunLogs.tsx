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

const makeHeader = (logs: LogEntry[], reverseSort: boolean) => {
  if (!logs.length) return 'No logs found';

  const timestamp = formatLogTimestamp(
    reverseSort ? logs[logs.length - 1].timestamp : logs[0].timestamp,
  );

  return `showing logs from ${
    reverseSort ? `now to ${timestamp}` : `${timestamp} to now`
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
    //turn useState into a promise so it finishes before refetch
    const stateActions = new Promise(() => {
      //select dropdown value
      stateFunction(value);
      //reset logs request
      setLogs([]);
      setToken('' as string);
    });

    stateActions.then(() => refetch());
  };

  React.useEffect(() => {
    if (isLoading) return;
    if (data?.logs?.length && data?.nextToken) {
      if (token)
        setLogs(
          reverseSort ? [...data.logs, ...logs] : [...logs, ...data.logs],
        );
      else setLogs(reverseSort ? data?.logs.reverse() : data?.logs);
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
          <MenuItem key="all" value={'all'}>
            all
          </MenuItem>
          <MenuItem key="info" value={'info'}>
            info
          </MenuItem>
          <MenuItem key="error" value={'error'}>
            error
          </MenuItem>
        </Select>
      </Flex>
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
