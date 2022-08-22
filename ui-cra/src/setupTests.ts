// still seems to work with vitest
// to generate snapshot with inline css rules
import 'jest-styled-components';

import matchers from '@testing-library/jest-dom/matchers';
import { expect } from 'vitest';

expect.extend(matchers);
