import type { JestConfigWithTsJest } from 'ts-jest';

const config: JestConfigWithTsJest = {
  preset: 'ts-jest/presets/js-with-babel',
  extensionsToTreatAsEsm: ['.ts', '.tsx'],
  testEnvironment: 'node',
  //   transformIgnorePatterns: ['/node_modules/'],
  //   moduleNameMapper: {
  //     '^(\\.{1,2}/.*)\\.js$': '$1',
  //   },
  //   transform: {
  //     '^.+\\.[tj]sx?$': [
  //       'ts-jest',
  //       {
  //         useESM: true,
  //       },
  //     ],
  //   },
};

export default config;
