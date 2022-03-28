import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { CallbackStateContextProvider } from '@weaveworks/weave-gitops';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import usePolicies from '../../contexts/policies';
import { Loader } from '../Loader';
import { PolicyTable } from './Table';

const Policies = () => {

  const { policies, loading } = usePolicies();

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Policies">
        <CallbackStateContextProvider>
          <SectionHeader
            className="count-header"
            path={[{ label: 'Policies', url: 'policies', count: policies.length }]}
          />
          <ContentWrapper>
            <Title>Policies</Title>
          </ContentWrapper>
        </CallbackStateContextProvider>
        {!loading ? (
          <div style={{ display: 'flex' }}>
           
              <ContentWrapper>
               
                <PolicyTable
                  count={policies.length}
                  policies={policies}
                />
              </ContentWrapper>
           
          </div>
        ) : (
          <ContentWrapper>
            <Loader />
          </ContentWrapper>
        )}
      </PageTemplate>
    </ThemeProvider>
  );
};

export default Policies;

