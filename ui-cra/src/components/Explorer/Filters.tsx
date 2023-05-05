import { Checkbox, FormControlLabel } from '@material-ui/core';
import _ from 'lodash';
import styled from 'styled-components';
import { Facet } from '../../api/query/query.pb';

export type FilterChangeHander = (vals: { [key: string]: boolean }) => void;

type Props = {
  className?: string;
  onFilterSelect?: FilterChangeHander;

  facets?: Facet[];
  state: { [key: string]: boolean };
};

function Filters({ className, onFilterSelect, facets, state }: Props) {
  const handleFilterChange = (field: string, key: string, value: boolean) => {
    const next = {
      ...state,
      [`+${field}:${key}`]: value,
    };

    onFilterSelect && onFilterSelect(next);
  };

  return (
    <div className={className}>
      {_.map(facets, f => {
        return (
          <div key={f.field}>
            <h3>{f.field}</h3>
            <ul style={{ listStyle: 'none' }}>
              {_.map(f.values, v => {
                return (
                  <li key={v}>
                    <FormControlLabel
                      label={v}
                      control={
                        <Checkbox
                          checked={state[v]}
                          onChange={e => {
                            handleFilterChange(
                              f.field as string,
                              v,
                              e.target.checked,
                            );
                          }}
                        />
                      }
                    />
                  </li>
                );
              })}
            </ul>
          </div>
        );
      })}
    </div>
  );
}

export default styled(Filters).attrs({ className: Filters.name })`
  ul {
    padding: 0px 8px;
  }

  h2 {
    margin-top: 0;
  }

  h3 {
    margin-bottom: 0;
  }
`;
