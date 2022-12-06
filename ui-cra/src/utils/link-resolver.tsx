// FIXME: remove this when core fixes requiring a linkResolver function
export const resolver = (path: string, params?: any) => {
  // Fix Kind as a path 
  return path.includes('/') ? path : '';
};
