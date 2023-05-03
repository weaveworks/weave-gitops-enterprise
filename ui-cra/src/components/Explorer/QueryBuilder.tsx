import {
  Box,
  CircularProgress,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  TextField,
} from '@material-ui/core';
import { Button, Flex, Input } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import * as React from 'react';
import styled from 'styled-components';

function noOp() {}

type Props = {
  className?: string;
  query: string;
  filters: { label: string; value: any }[];
  selectedFilter: string;
  disabled?: boolean;
  placeholder?: string;
  onChange: (val: string) => void;
  onFilterSelect: (val: string) => void;
  onSubmit: (val: string) => void;
  onBlur?: () => void;
  busy?: boolean;
  hideTextInput?: boolean;
};

function QueryBuilder({
  query,
  filters,
  selectedFilter,
  disabled,
  className,
  placeholder,
  onChange,
  onSubmit,
  onBlur = noOp,
  onFilterSelect,
  busy,
  hideTextInput,
}: Props) {
  const inputRef = React.useRef<HTMLInputElement>(null);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange(e.target.value);
  };

  const handleFocus = () => {};

  const handleFilterChange = (
    ev: React.ChangeEvent<{ name?: string; value: any }>,
  ) => {
    if (inputRef.current) {
      inputRef.current.focus();
    }
    onChange('');
    onFilterSelect(ev.target.value);
  };

  return (
    <div className={className}>
      <Flex align>
        <Box marginRight={1}>
          {!hideTextInput && (
            <form
              onSubmit={e => {
                e.preventDefault();
                onSubmit(query);
              }}
            >
              <Flex>
                <FormControl>
                  <TextField
                    placeholder={placeholder}
                    style={{ minWidth: 360 }}
                    variant="outlined"
                    onChange={handleInputChange}
                    value={query}
                    // onKeyDown={handleInputKeyPress}
                    onBlur={onBlur}
                    onFocus={handleFocus}
                    inputRef={inputRef}
                    disabled={disabled}
                    onSubmit={e => {}}
                  />
                </FormControl>
                <Box marginLeft={1}>
                  <Button style={{ height: 56 }} type="submit">
                    Submit
                  </Button>
                </Box>
              </Flex>
            </form>
          )}
        </Box>
        {!_.isEmpty(filters) && (
          <Box>
            <Flex align>
              <FormControl
                variant="outlined"
                style={{ minWidth: 240, marginRight: 16 }}
              >
                <InputLabel id="demo-simple-select-label">Filters</InputLabel>
                <Select
                  label="Filters"
                  placeholder="Filters"
                  onChange={handleFilterChange}
                  value={selectedFilter}
                >
                  <MenuItem value="">
                    <em>None</em>
                  </MenuItem>
                  {_.map(filters, filter => (
                    <MenuItem key={filter.label} value={filter.value}>
                      {filter.label}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Flex>
          </Box>
        )}
        {busy && (
          <Box marginLeft={2}>
            <CircularProgress size={24} />
          </Box>
        )}
      </Flex>
    </div>
  );
}

export default styled(QueryBuilder).attrs({ className: QueryBuilder.name })`
  position: relative;

  ${Input} {
    flex: 2;
    width: 100%;
    input {
      padding: 0 8px;
      width: 100%;
    }
    input:focus {
      outline: none;
    }
  }
`;
