import {
  ChipGroup,
  FilterDialog,
  Flex,
  formStateToFilters,
  Icon,
  IconButton,
  IconType,
  initialFormState,
  SearchField,
} from '@weaveworks/weave-gitops';
import {
  DialogFormState,
  FilterConfig,
} from '@weaveworks/weave-gitops/ui/components/FilterDialog';
import _ from 'lodash';
import * as React from 'react';
import styled from 'styled-components';
import { Condition } from '../../capi-server/capi_server.pb';
import DataTable, { Field } from './DataTable';
// import FilterDialog, {
//   DialogFormState,
//   FilterConfig,
//   formStateToFilters,
//   initialFormState,
// } from './FilterDialog';

type Props = {
  className?: string;
  fields: Field[];
  rows: any[];
  filters: FilterConfig;
  dialogOpen?: boolean;
  onDialogClose?: () => void;
};

export function computeReady(conditions: Condition[]): boolean {
  const ready =
    _.find(conditions, { type: 'Ready' }) ||
    // Deployment conditions work slightly differently;
    // they show "Available" instead of 'Ready'
    _.find(conditions, { type: 'Available' });

  return ready?.status == 'True';
}

export function filterConfigForString(rows: any, key: string) {
  const typeFilterConfig = _.reduce(
    rows,
    (r, v) => {
      const t = v[key];

      if (!_.includes(r, t)) {
        // @ts-ignore
        r.push(t);
      }

      return r;
    },
    [],
  );

  return { [key]: typeFilterConfig };
}

export function filterConfigForStatus(rows: any) {
  const statusFilterConfig = _.reduce(
    rows,
    (r, v) => {
      let t;
      if (v.suspended) t = 'Suspended';
      else if (computeReady(v.conditions)) t = 'Ready';
      else t = 'Not Ready';
      if (!_.includes(r, t)) {
        // @ts-ignore
        r.push(t);
      }
      return r;
    },
    [],
  );

  return { status: statusFilterConfig };
}

export function filterRows<T>(rows: T[], filters: FilterConfig) {
  if (_.keys(filters).length === 0) {
    return rows;
  }

  return _.filter(rows, r => {
    let ok = false;

    _.each(filters, (vals, key) => {
      let value;
      //status
      if (key === 'status') {
        // @ts-ignore

        if (r['suspended']) value = 'Suspended';
        // @ts-ignore
        else if (computeReady(r['conditions'])) value = 'Ready';
        else value = 'Not Ready';
      }
      //string
      // @ts-ignore
      else value = r[key];

      if (_.includes(vals, value)) {
        ok = true;
      }
    });

    return ok;
  });
}

export function filterText(
  // @ts-ignore
  rows,
  fields: Field[],
  textFilters: State['textFilters'],
) {
  if (textFilters.length === 0) {
    return rows;
  }

  return _.filter(rows, row => {
    let matches = false;
    for (const colName in row) {
      const value = row[colName];

      const field = _.find(fields, f => {
        if (typeof f.value === 'string') {
          return f.value === colName;
        }

        if (f.sortValue) {
          return f.sortValue(row) === value;
        }
      });

      // @ts-ignore
      if (!field || !field.textSearchable) {
        continue;
      }

      // This allows us to look for a fragment in the string.
      // For example, if the user searches for "pod", the "podinfo" kustomization should be returned.
      for (const filterString of textFilters) {
        if (_.includes(value, filterString)) {
          matches = true;
        }
      }
    }

    return matches;
  });
}

export function toPairs(state: State): string[] {
  const result = _.map(state.formState, (val, key) => (val ? key : null));
  const out = _.compact(result);
  return _.concat(out, state.textFilters);
}

export type State = {
  filters: FilterConfig;
  formState: DialogFormState;
  textFilters: string[];
};

function FilterableTable({
  className,
  fields,
  rows,
  filters,
  dialogOpen,
}: Props) {
  const [filterDialogOpen, setFilterDialogOpen] = React.useState(dialogOpen);
  const [filterState, setFilterState] = React.useState<State>({
    filters,
    formState: initialFormState(filters),
    textFilters: [],
  });
  let filtered = filterRows(rows, filterState.filters);
  console.log(filtered);

  filtered = filterText(filtered, fields, filterState.textFilters);
  const chips = toPairs(filterState);

  const handleChipRemove = (chips: string[]) => {
    const next = {
      ...filterState,
    };

    _.each(chips, chip => {
      next.formState[chip] = false;
    });

    const filters = formStateToFilters(next.formState);

    const textFilters = _.filter(next.textFilters, f => !_.includes(chips, f));

    setFilterState({ formState: next.formState, filters, textFilters });
  };

  const handleTextSearchSubmit = (val: string) => {
    setFilterState({
      ...filterState,
      textFilters: _.uniq(_.concat(filterState.textFilters, val)),
    });
  };

  const handleClearAll = () => {
    setFilterState({
      filters: {},
      formState: initialFormState(filters),
      textFilters: [],
    });
  };

  const handleFilterSelect = (filters: any, formState: any) => {
    setFilterState({ ...filterState, filters, formState });
  };

  console.log(filtered);

  return (
    <Flex className={className} wide tall column>
      <Flex wide align>
        <ChipGroup
          chips={chips}
          onChipRemove={handleChipRemove}
          onClearAll={handleClearAll}
        />
        <Flex align wide end>
          <SearchField onSubmit={handleTextSearchSubmit} />
          <IconButton
            onClick={() => setFilterDialogOpen(!filterDialogOpen)}
            className={className}
            variant={filterDialogOpen ? 'contained' : 'text'}
            color="inherit"
          >
            <Icon type={IconType.FilterIcon} size="medium" color="neutral30" />
          </IconButton>
        </Flex>
      </Flex>
      <Flex wide tall>
        <DataTable className={className} fields={fields} rows={filtered} />
        <FilterDialog
          onFilterSelect={handleFilterSelect}
          filterList={filters}
          formState={filterState.formState}
          open={filterDialogOpen}
        />
      </Flex>
    </Flex>
  );
}

export default styled(FilterableTable).attrs({
  className: FilterableTable.name,
})``;
