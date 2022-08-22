import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ContentWrapper } from '../../Layout/ContentWrapper';

import HeaderSection from './HeaderSection';
import ParametersSection from './ParametersSection';
import { useGetPolicyDetails } from '../../../contexts/PolicyViolations';
import { Alert } from '@material-ui/lab';
import { LoadingPage } from '@weaveworks/weave-gitops';

const PolicyDetails = ({
  clusterName,
  id,
}: {
  clusterName: string;
  id: string;
}) => {
  const { data, error, isLoading } = useGetPolicyDetails({
    clusterName,
    policyName: id,
  });
  const policy = data?.policy;
  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Policies">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Policies', url: '/policies' },
            { label: data?.policy?.name || '' },
          ]}
        />
        <ContentWrapper>
          {isLoading && <LoadingPage />}
          {error && <Alert severity="error">{error.message}</Alert>}
          {data?.policy && (
            <>
              <HeaderSection
                id={policy?.id}
                clusterName={policy?.clusterName}
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
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PolicyDetails;
