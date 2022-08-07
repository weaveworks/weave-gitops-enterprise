import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../muiTheme';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { FieldsType, PolicyViolationsList } from './Table';
import useClusters from './../../contexts/Clusters';
import { useListPolicyValidations } from '../../contexts/PolicyViolations';

const PoliciesViolations = () => {
  const clustersCount = useClusters().count;
  const { data } = useListPolicyValidations({});

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Violation Log">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Clusters', url: '/clusters', count: clustersCount },
            {
              label: 'Violation Log',
              url: 'violations',
              count: data?.total,
            },
          ]}
        />
        <ContentWrapper errors={data?.errors}>
          <PolicyViolationsList req={{}} tableType={ FieldsType.policy} />
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PoliciesViolations;
