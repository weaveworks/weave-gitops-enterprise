import { FormControl } from '@material-ui/core';
import { Flex, Input } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { useState } from 'react';
import styled from 'styled-components';
import { QueryState } from './hooks';

type Props = {
  className?: string;

  queryState: QueryState;
  onTextInputChange?: (value: string) => void;
};

const debouncedInputHandler = _.debounce((fn, val) => {
  fn(val);
}, 500);

function QueryInput({
  className,
  queryState: state,
  onTextInputChange,
}: Props) {
  const [textInput, setTextInput] = useState(state.terms || '');

  const handleTextInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    setTextInput(e.target.value);

    debouncedInputHandler(onTextInputChange, e.target.value);
  };

  return (
    <Flex className={className} wide>
      <FormControl>
        <Input value={textInput} onChange={handleTextInput} />
      </FormControl>
    </Flex>
  );
}

export default styled(QueryInput).attrs({ className: QueryInput.name })`
  position: relative;
`;
