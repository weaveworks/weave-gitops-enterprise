import { Box, IconButton } from '@material-ui/core';
import { Flex, Icon, IconType, ThemeTypes } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { QueryState } from './hooks';

type Props = {
  className?: string;
  queryState: QueryState;
  setQueryState: (state: QueryState) => void;
  count: number;
};

function PaginationControls({
  className,
  queryState,
  setQueryState,
  count,
}: Props) {
  const handlePageForward = () => {
    setQueryState({
      ...queryState,
      offset: queryState.offset + queryState.limit,
    });
  };

  const handlePageBack = () => {
    setQueryState({
      ...queryState,
      offset: Math.max(0, queryState.offset - queryState.limit),
    });
  };

  return (
    <Flex className={className} wide center>
      <Box p={2}>
        <IconButton disabled={queryState.offset === 0} onClick={handlePageBack}>
          <Icon size={24} type={IconType.NavigateBeforeIcon} />
        </IconButton>
        <IconButton
          disabled={count < queryState.limit}
          onClick={handlePageForward}
        >
          <Icon size={24} type={IconType.NavigateNextIcon} />
        </IconButton>
      </Box>
    </Flex>
  );
}

export default styled(PaginationControls).attrs({
  className: PaginationControls.name,
})`
  ${Icon} .MuiSvgIcon-root {
    color: ${props =>
      props.theme.mode === ThemeTypes.Dark
        ? props.theme.colors.white
        : 'unset'};
  }
`;
