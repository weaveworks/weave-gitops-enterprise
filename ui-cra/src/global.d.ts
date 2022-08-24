import type { TestingLibraryMatchers } from '@testing-library/jest-dom/matchers';

// After extending the vitest matchers with testing-library in src/setupTests.ts
// (to additionally provide `toHaveTextContent` etc), we also need to patch up the types.
//
// When testing-library hopefully eventually releases a vitest matching lib we won't have to do this anymore.
//
// https://github.com/testing-library/jest-dom/issues/427#issuecomment-1110985202
//
declare global {
  namespace jest {
    type Matchers<R = void, T = {}> = TestingLibraryMatchers<
      typeof expect.stringContaining,
      R
    >;
  }
}
