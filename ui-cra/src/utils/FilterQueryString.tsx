export const toFilterQueryString = (
  filters: { key: string; value: string }[],
) => {
  const filtersValues = encodeURIComponent(
    filters.map(filter => `${filter.key}: ${filter.value}`).join('_'),
  );
  return filtersValues;
};
