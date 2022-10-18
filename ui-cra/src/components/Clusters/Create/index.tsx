import Grid from '@material-ui/core/Grid';
import { useParams } from 'react-router-dom';
import useTemplates from '../../../hooks/templates';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import ClusterForm from '../Form';

const CreateClusterPage = () => {
  const { templateName } = useParams<{ templateName: string }>();
  const { getTemplate, isLoading } = useTemplates();
  return (
    <PageTemplate
      documentTitle="Create new cluster"
      path={[
        { label: 'Templates', url: '/templates' },
        { label: 'Create new cluster' },
      ]}
    >
      <ContentWrapper loading={isLoading}>
        <Grid container>
          <Grid item xs={12} sm={10} md={10} lg={8}>
            <Title>Create new cluster with template</Title>
          </Grid>
          <ClusterForm template={getTemplate(templateName)} />
        </Grid>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default CreateClusterPage;
