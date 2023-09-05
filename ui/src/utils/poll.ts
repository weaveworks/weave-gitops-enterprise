import { PureComponent } from 'react';
import moment from 'moment';
import { map, find, keys, values, zipObject } from 'lodash';

import { request } from '../utils/request';
import { UnixTimestampMilliseconds } from '../types/global';

type QueriesById<ResponsesById> = {
  [queryId in keyof ResponsesById]: RequestInfo;
};

interface Props<ResponsesById> {
  queriesById: QueriesById<ResponsesById>;
  requestOptions?: RequestInit;
  intervalMs: number;
  children: (props: PollState<ResponsesById>) => JSX.Element;
}

// Note: this breaks from the convention of simply calling it `State`, since the name of this interface is exposed to the user through the render prop (see Props.children).
export interface PollState<ResponsesById> {
  error: string | null;
  isLoading: boolean;
  responsesById: ResponsesById;
  timestamp: UnixTimestampMilliseconds;
}

export class Poll<ResponsesById> extends PureComponent<
  Props<ResponsesById>,
  PollState<ResponsesById>
> {
  static defaultProps = {
    intervalMs: 15000,
    requestOptions: {},
  };

  state: PollState<ResponsesById> = {
    error: null,
    isLoading: true, // we initalize to true because otherwise the first data sent to the child component will have `isLoading: false` _as well as_ having `null` response data (because the response hasn't be sent yet).
    responsesById: {} as ResponsesById,
    timestamp: moment().valueOf(),
  };

  timeoutId: NodeJS.Timeout | null = null;

  componentDidMount() {
    this.poll();
  }

  componentWillUnmount() {
    if (this.timeoutId !== null) {
      clearTimeout(this.timeoutId);
    }
  }

  doParallelRequests = () => {
    const { requestOptions, queriesById } = this.props;
    const requestInfo = values<QueriesById<ResponsesById>>(queriesById);
    return new Promise<ResponsesById>((resolve, reject) => {
      // Make requests for all the queries at the same time.
      Promise.all(
        requestInfo.map((query: any) => request('GET', query, requestOptions)),
      )
        .then(responses => {
          // If any the queries responds with an error, reject the whole promise ...
          const errorResponse = find(responses, { status: 'error' });
          if (errorResponse) {
            // TODO(dimitri): fix promise rejects not return `Error`s
            // eslint-disable-next-line prefer-promise-reject-errors
            reject(`${errorResponse.errorType}: ${errorResponse.error}`);
          }
          // ... otherwise resolve, mapping the results back to the query keys.
          const responsesById = zipObject(
            keys(queriesById),
            // TODO(fbarl): Consider wrapping workspace listing in 'data' unify the API format.
            // See https://github.com/weaveworks/wks/issues/926.
            map(responses, response =>
              response.data ? response.data : response,
            ),
          ) as ResponsesById;
          resolve(responsesById);
        })
        .catch(reason => {
          // Propagate the standard rejection reasons.
          reject(reason.message);
        });
    });
  };

  poll = () => {
    const { intervalMs } = this.props;
    this.setState({
      isLoading: true,
      timestamp: moment().valueOf(),
    });
    this.doParallelRequests()
      .then(responsesById => {
        this.setState({
          error: null,
          isLoading: false,
          responsesById,
        });
      })
      .catch(error => {
        // NOTE: this means that the semantics of this component are that if an error returns it will still continue to return the (last) stale data in responsesById since responsesById is not also cleared here.
        this.setState({ error, isLoading: false });
      })
      .finally(() => {
        // Keep polling regardless of whether the last request succeeded or not.
        this.timeoutId = setTimeout(this.poll, intervalMs);
      });
  };

  render() {
    const { children } = this.props;
    return children(this.state);
  }
}
