import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { PolicyTable } from './Table';
import { PolicyService } from './PolicyService';
import { useCallback, useState } from 'react';
import { ListPoliciesResponse } from '../../capi-server/capi_server.pb';
import LoadingError from '../LoadingError';

const Policies = () => {
  const [count, setCount] = useState<number>(0);
  const [payload, setPayload] = useState<any>({ page: 1, limit: 25 });

  // I used callback here because I need to pass the payload to the API call as well as the setter function to update the payload in the state object (payload) and I don't want to pass the payload to the API call as well.
  // I could have used useState and setState but I wanted to keep the code as simple as possible.
  const fetchPoliciesAPI = useCallback(() => {
    return PolicyService.listPolicies(payload).then(
      (res: ListPoliciesResponse | any) => {
        !!res && setCount(res.total);
        return res;
      },
    );
  }, [payload]);

  // Update payload on page change for next page request to work properly with pagination component in PolicyTable component below
  // const updatePayload = (payload: any) => {
  //   setPayload(payload);
  // };

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Policies">
        <SectionHeader
          className="count-header"
          path={[{ label: 'Policies', url: 'policies', count: count }]}
        />
        <ContentWrapper>
          <Title>Policies</Title>
          <LoadingError fetchFn={fetchPoliciesAPI}>
            {({ value }: { value: ListPoliciesResponse }) => (
              <>
                {value.total && value.total > 0 ? (
                  <PolicyTable policies={value.policies} />
                ) : (
                  <div>No data to display</div>
                )}
              </>
            )}
          </LoadingError>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default Policies;
