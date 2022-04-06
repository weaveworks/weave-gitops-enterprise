import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { CallbackStateContextProvider } from '@weaveworks/weave-gitops';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';

import { PolicyService } from './../PolicyService';
import { useState } from 'react';
import LoadingError from '../../LoadingError';
import HeaderSection from './HeaderSection';
import { useParams } from 'react-router-dom';
import { GetPolicyResponse } from '../../../capi-server/capi_server.pb';

const PolicyDetails = () => {
  const { id } = useParams<{ id: string }>();
  const [name, setName] = useState('');
  const fetchPoliciesAPI = () =>
    PolicyService.getPolicyById(id).then((res: GetPolicyResponse) => {
      res.policy?.name && setName(res.policy.name);
      return res;
    });

  const [fetchPolicyById] = useState(() => fetchPoliciesAPI);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Policies">
        <CallbackStateContextProvider>
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
              {({ value: { policy } }: { value: GetPolicyResponse }) => (
                <>
                  <HeaderSection
                    id={policy?.id}
                    tags={policy?.tags}
                    severity={policy?.severity}
                    category={policy?.category}
                    targets={policy?.targets}
                    description={policy?.description}
                    howToSolve={policy?.howToSolve}
                  ></HeaderSection>
                </>
              )}
            </LoadingError>
          </ContentWrapper>
        </CallbackStateContextProvider>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PolicyDetails;
