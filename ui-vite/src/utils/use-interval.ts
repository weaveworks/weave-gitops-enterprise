import { useEffect, useRef } from 'react';
import { noop } from 'lodash';

// From https://overreacted.io/making-setinterval-declarative-with-react-hooks/
export function useInterval(
  callback: () => void,
  delay: number | null,
  runInitially?: boolean,
  deps?: React.DependencyList,
) {
  const savedCallback = useRef(noop);

  // Remember the latest callback.
  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  useEffect(() => {
    if (runInitially && savedCallback.current) {
      savedCallback.current();
    }
    // the `callback` dependency is not included here because we only want to run this effect on startup, or when `runInitially` changes.
    // See https://github.com/weaveworks/wkp-ui/pull/211 for more context.
  }, [runInitially, ...(deps ?? [])]); // eslint-disable-line react-hooks/exhaustive-deps

  // Set up the interval.
  useEffect(() => {
    function tick() {
      savedCallback.current();
    }
    if (delay !== null) {
      const id = setInterval(tick, delay);
      return () => clearInterval(id);
    }
    return undefined;
  }, [delay, ...(deps ?? [])]); // eslint-disable-line react-hooks/exhaustive-deps
}
