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
import ViolationDetails from './ViolationDetails';

const PolicyViolationDetails = () => {
  const { id } = useParams<{ id: string }>();
  const [name, setName] = useState('');
  const fetchPoliciesViolationsAPI = () =>
    PolicyService.getPolicyViolationById(id).then(
      (res: GetPolicyValidationResponse) => {
        res.violation?.name && setName(res.violation.name);
        return res;
      },
    );

  const [fetchPolicyViolationById] = useState(() => fetchPoliciesViolationsAPI);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Violations Logs">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Violation logs', url: '/clusters/violations' },
            { label: name, url: 'policy-violation-details' },
          ]}
        />
        <ContentWrapper>
          <Title>{name}</Title>
          <LoadingError fetchFn={fetchPolicyViolationById}>
            {({
              value: { violation },
            }: {
              value: GetPolicyValidationResponse;
            }) => (
              <>
                <ViolationDetails violation={violation}></ViolationDetails>
              </>
            )}
          </LoadingError>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PolicyViolationDetails;
