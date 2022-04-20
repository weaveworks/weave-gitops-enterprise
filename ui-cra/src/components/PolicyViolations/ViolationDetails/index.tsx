import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';

import { useState } from 'react';
import LoadingError from '../../LoadingError';
import { useParams } from 'react-router-dom';
import { GetPolicyValidationResponse } from '../../../capi-server/capi_server.pb';
import { PolicyService } from '../../Policies/PolicyService';

const PolicyViolationDetails = () => {
  const { id } = useParams<{ id: string }>();
  const [name, setName] = useState('');
  const fetchPoliciesAPI = () =>
    PolicyService.getPolicyViolationById(id).then((res: GetPolicyValidationResponse) => {
      res.violation?.name && setName(res.violation.name);
      return res;
    });

  const [fetchPolicyById] = useState(() => fetchPoliciesAPI);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Policies">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Policies', url: '/policies' },
            { label: name, url: 'policy-details' },
          ]}
        />
        <ContentWrapper>
          <Title>{name}</Title>
          <LoadingError fetchFn={fetchPolicyById}>
            {({ value: { policy } }: { value: GetPolicyValidationResponse }) => (
              <>
               
               
              </>
            )}
          </LoadingError>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PolicyViolationDetails;
