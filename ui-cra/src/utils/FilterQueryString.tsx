export const toFilterQueryString = (filters: { [key: string]: string }[]) => {
  const filtersValues = encodeURIComponent(
    filters
      .map(
        filter =>
          `${Object.keys(filter)[0]}: ${Object.values(filter)[0] || ''}`,
      )
      .join('_'),
  );
  return filtersValues;
};
