import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { CallbackStateContextProvider } from '@weaveworks/weave-gitops';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';

import { PolicyService } from './../PolicyService';
import { useState } from 'react';
import { Policy } from '../../../types/custom';
import LoadingError from '../../LoadingError';
import { useParams } from 'react-router-dom';

interface IPolicyDetailsResponse {
  policy: Policy;
}

const PolicyDetails = () => {
const { id } = useParams<{id:string}>();
const [name,setName] = useState('');
const fetchPoliciesAPI = () => PolicyService.getPolicyById(id).then((res: any) => {setName(res.policy.name); return res;});




  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Policies">
        <CallbackStateContextProvider>
          <SectionHeader
            className="count-header"
            path={[{ label: 'Policies', url: 'policies'}, { label: name, url: 'policy-details' }]}
          />
          <ContentWrapper>
            <Title>{name}</Title>
            <LoadingError fetchFn={fetchPoliciesAPI}>
              {({ value }: { value: IPolicyDetailsResponse }) => ( 
                <>
                {console.log(value)}
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
