import {
  Box,
  CircularProgress,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  TextField,
} from '@material-ui/core';
import { ChipGroup, Flex, Input } from '@weaveworks/weave-gitops';
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
  pinnedTerms: string[];
  onChange: (val: string, pinnedTerms: string[]) => void;
  onFilterSelect: (val: string) => void;
  onPin: (val: string[]) => void;
  onBlur?: () => void;
  busy?: boolean;
  hideTextInput?: boolean;
};

function QueryBuilder({
  query,
  pinnedTerms,
  filters,
  selectedFilter,
  disabled,
  className,
  placeholder,
  onChange,
  onPin,
  onBlur = noOp,
  onFilterSelect,
  busy,
  hideTextInput,
}: Props) {
  const inputRef = React.useRef<HTMLInputElement>(null);

  const handleAddSearchTerm = (value: string) => {
    let nextPinnedTerms = pinnedTerms;
    // only push unique values
    if (!_.includes(nextPinnedTerms, value)) {
      nextPinnedTerms = [...nextPinnedTerms, value];
      onPin(nextPinnedTerms);
    }
    onChange('', nextPinnedTerms);
  };

  const handleRemoveSearchTerm = (value: string[]) => {
    const nextPinnedTerms = _.without(pinnedTerms, value[0]);
    onChange(query, nextPinnedTerms);
  };

  const handleRemoveAll = () => {
    onChange('', []);
    onFilterSelect('');
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange(e.target.value, pinnedTerms);
  };

  const handleInputKeyPress = (ev: React.KeyboardEvent<HTMLInputElement>) => {
    if (ev.key === 'Enter' && query.length > 0) {
      ev.preventDefault();
      handleAddSearchTerm(query);
    } else if (ev.key === 'Backspace' && query === '') {
      ev.preventDefault();
      const term = _.last(pinnedTerms);
      if (term) {
        // Allow the user to edit the text of the last term instead of removing the whole thing.
        handleRemoveSearchTerm([term]);
      }
    }
  };

  const handleFocus = () => {};

  const handleFilterChange = (
    ev: React.ChangeEvent<{ name?: string; value: any }>,
  ) => {
    if (inputRef.current) {
      inputRef.current.focus();
    }

    onFilterSelect(ev.target.value);
  };

  return (
    <div className={className}>
      <Box marginBottom={1}>
        <ChipGroup
          chips={pinnedTerms}
          onChipRemove={handleRemoveSearchTerm}
          onClearAll={handleRemoveAll}
        />
      </Box>

      <Flex align>
        <Box marginRight={1}>
          {!hideTextInput && (
            <TextField
              placeholder={placeholder}
              style={{ minWidth: 360 }}
              variant="outlined"
              onChange={handleInputChange}
              value={query}
              onKeyDown={handleInputKeyPress}
              onBlur={onBlur}
              onFocus={handleFocus}
              inputRef={inputRef}
              disabled={disabled}
            />
          )}
        </Box>
        {!_.isEmpty(filters) && (
          <Box>
            <FormControl variant="outlined" style={{ minWidth: 124 }}>
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
