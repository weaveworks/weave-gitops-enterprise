import { useReadQueryState, useSetQueryState } from './hooks';
import { ChipGroup } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

type Props = {
  className?: string;
};

function QueryStateChips({ className }: Props) {
  const queryState = useReadQueryState();
  const setQueryState = useSetQueryState();

  const chips = [...queryState.filters];

  if (queryState.terms) {
    chips.push(`terms:${queryState.terms}`);
  }

  const handleClearAll = () => {
    setQueryState({
      ...queryState,
      filters: [],
      terms: '',
      offset: 0,
    });
  };

  const handleChipRemove = (chips: string[]) => {
    for (const chip of chips) {
      if (chip.includes('terms:')) {
        setQueryState({
          ...queryState,
          terms: '',
        });
        return;
      }
      setQueryState({
        ...queryState,
        filters: queryState.filters.filter(c => c !== chip),
      });
    }
  };

  return (
    <div className={className}>
      <ChipGroup
        chips={chips}
        onChipRemove={handleChipRemove}
        onClearAll={handleClearAll}
      />
    </div>
  );
}

export default styled(QueryStateChips).attrs({
  className: QueryStateChips.name,
})``;
