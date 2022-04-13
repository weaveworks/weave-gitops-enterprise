import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';

import { PolicyService } from './../PolicyService';
import { useCallback, useState } from 'react';
import LoadingError from '../../LoadingError';
import HeaderSection from './HeaderSection';
import { useParams } from 'react-router-dom';
import { GetPolicyResponse } from '../../../capi-server/capi_server.pb';
import ParametersSection from './ParametersSection';

const PolicyDetails = () => {
  const { id } = useParams<{ id: string }>();
  const [name, setName] = useState<string>('');

  const fetchPoliciesAPI = useCallback(() => {
    return PolicyService.getPolicyById(id).then((res: GetPolicyResponse) => {
      setName(res.policy?.name || '');
      return res;
    });
  }, [id]);

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
          <LoadingError fetchFn={fetchPoliciesAPI}>
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
                  code={policy?.code}
                ></HeaderSection>
                <ParametersSection
                  parameters={policy?.parameters}
                ></ParametersSection>
              </>
            )}
          </LoadingError>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PolicyDetails;
