import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { CallbackStateContextProvider } from '@weaveworks/weave-gitops';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import usePolicies from '../../contexts/Policies';
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
          {!loading ? (
          <div style={{ display: 'flex' }}>
           
              <ContentWrapper>
              <Title>Policies</Title>
               
                <PolicyTable
                  count={policies.length}
                  policies={policies}
                />
              </ContentWrapper>
           
          </div>
        ) : (
          <ContentWrapper>
            <Title>Policies</Title>
            <Loader />
          </ContentWrapper>
        )}
        </CallbackStateContextProvider>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default Policies;

