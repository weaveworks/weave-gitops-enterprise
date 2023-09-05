import { Checkbox, FormControlLabel } from '@material-ui/core';
import _ from 'lodash';
import styled from 'styled-components';
import { Facet } from '../../api/query/query.pb';
import { useReadQueryState, useSetQueryState } from './hooks';

type Props = {
  className?: string;

  facets?: Facet[];
};

function Filters({ className, facets }: Props) {
  const queryState = useReadQueryState();
  const setState = useSetQueryState();

  const handleFilterChange = (field: string, key: string, value: boolean) => {
    const record = `${field}:${key}`;

    const existing = _.find(queryState.filters, f => f === record);

    // Reset the offset when filters change.
    const offset = 0;

    const filters =
      existing && !value
        ? _.filter(queryState.filters, f => f !== record)
        : _.concat(queryState.filters, record);

    setState({
      ...queryState,
      filters,
      offset,
    });
  };

  const filterState = _.reduce(
    queryState.filters,
    (result, f) => {
      const re = /(.+?):(.*)/g;

      const matches = re.exec(f);

      if (matches) {
        const [, key, value] = matches;

        result[`${key.replace('+', '')}:${value}`] = true;
      }

      return result;
    },
    {} as { [key: string]: boolean },
  );

  return (
    <div className={className}>
      {_.map(facets, f => {
        return (
          <div key={f.field}>
            <h3>{_.capitalize(f.field)}</h3>
            <ul style={{ listStyle: 'none' }}>
              {_.map(f.values, v => {
                const key = `${f.field}:${v}`;

                return (
                  <li key={v}>
                    <FormControlLabel
                      label={v}
                      control={
                        <Checkbox
                          // Leaving this as uncontrolled for now.
                          // URL state is proving to be a problem and
                          // the PR is already very sprawling.
                          checked={filterState[key] || false}
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
