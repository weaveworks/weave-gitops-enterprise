import { MenuItem, Select } from '@material-ui/core';
import { Input } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import * as React from 'react';
import styled from 'styled-components';
import MultiSelectDropdown from '../MultiSelectDropdown';

const SearchInput = styled.div`
  display: flex;
  flex-wrap: wrap;
  flex: 3;
  width: 100%;
`;

const TermsContainer = styled.ul<{ disabled: boolean }>`
  list-style: none;
  display: flex;
  margin: 0;
  padding: 0;
  flex-wrap: wrap;

  ${props => props.disabled && 'opacity: 0.75;'};
`;

class SearchTerm extends React.PureComponent<any> {
  handleRemove = () => {
    this.props.onRemove(this.props.term, this.props.label);
  };

  render() {
    const { className, term, label } = this.props;
    return (
      <li className={`${className} search-term`}>
        <div className="search-term-text">{label || term}</div>
        <i onClick={this.handleRemove} className="fa fa-times remove-term" />
      </li>
    );
  }
}

function noOp() {}

type Props = {
  className?: string;
  query: string;
  filters: { label: string; value: any }[];
  disabled: boolean;
  placeholder?: string;
  pinnedTerms: string[];
  onChange: (val: string, pinnedTerms: string[]) => void;
  onFilterSelect: (val: string) => void;
  onPin: (val: string[]) => void;
  onBlur?: () => void;
};

function QueryBuilder({
  query,
  pinnedTerms,
  filters,
  disabled,
  className,
  placeholder,
  onChange,
  onPin,
  onBlur = noOp,
  onFilterSelect,
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

  const handleRemoveSearchTerm = (value: string) => {
    const nextPinnedTerms = _.without(pinnedTerms, value);
    onChange(query, nextPinnedTerms);
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
        handleRemoveSearchTerm(term);
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
      <SearchInput>
        <TermsContainer disabled={disabled}>
          {_.map(pinnedTerms, term => (
            <SearchTerm
              key={term}
              term={term}
              onRemove={handleRemoveSearchTerm}
            />
          ))}
        </TermsContainer>
        <Input
          onChange={handleInputChange}
          value={query}
          onKeyDown={handleInputKeyPress}
          onBlur={onBlur}
          onFocus={handleFocus}
          inputRef={inputRef}
          // placeholder={pinnedTerms.length === 0 ? placeholder : null}
          disabled={disabled}
        />
      </SearchInput>

      {!_.isEmpty(filters) && (
        <Select placeholder="Filters" onChange={handleFilterChange}>
          {_.map(filters, filter => (
            <MenuItem key={filter.label} value={filter.value}>
              {filter.label}
            </MenuItem>
          ))}
        </Select>
      )}
    </div>
  );
}

export default styled(QueryBuilder).attrs({ className: QueryBuilder.name })`
  position: relative;
  display: flex;

  align-items: center;

  div,
  input {
    border: 0;
  }

  ${MultiSelectDropdown} {
    flex: 1;
    line-height: 36px;

    .dropdown-popover {
      width: auto;
    }
  }

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
