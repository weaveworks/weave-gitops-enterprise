import { useReadQueryState, useSetQueryState } from './hooks';
import { FormControl } from '@material-ui/core';
import { Flex, Icon, IconType, Input } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { useEffect, useState } from 'react';
import styled from 'styled-components';

type Props = {
  className?: string;
};

const debouncedInputHandler = _.debounce((fn, val) => {
  fn(val);
}, 500);

function QueryInput({ className }: Props) {
  const queryState = useReadQueryState();
  const setQueryState = useSetQueryState();
  const [textInput, setTextInput] = useState(queryState.terms || '');

  useEffect(() => {
    setTextInput(queryState.terms || '');
  }, [queryState.terms]);

  const handleTextInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    setTextInput(e.target.value);

    debouncedInputHandler(
      (val: string) => setQueryState({ ...queryState, terms: val }),
      e.target.value,
    );
  };

  return (
    <Flex className={className} wide>
      <Flex align>
        <Icon size="normal" type={IconType.SearchIcon} />
        <FormControl>
          <Input
            placeholder="Search"
            value={textInput}
            onChange={handleTextInput}
          />
        </FormControl>
      </Flex>
    </Flex>
  );
}

export default styled(QueryInput).attrs({ className: QueryInput.name })`
  position: relative;
`;
