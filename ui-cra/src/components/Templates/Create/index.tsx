import Grid from '@material-ui/core/Grid';
import { useLocation, useParams } from 'react-router-dom';
import useTemplates from '../../../hooks/templates';
import { NotificationsWrapper, Title } from '../../Layout/NotificationsWrapper';
import ResourceForm from '../Form';
import { Page } from '../../Layout/App';

const CreateResourcePage = () => {
  // const { templateName } = useParams<{ templateName: string }>();
  const { search } = useLocation();
  const templateName = new URLSearchParams(search).get('name') as string;
  const { getTemplate, isLoading } = useTemplates();
  return (
    <Page
      loading={isLoading}
      path={[
        { label: 'Templates', url: '/templates' },
        { label: 'Create new resource' },
      ]}
    >
      <NotificationsWrapper>
        <Grid container>
          <Grid item xs={12} sm={10} md={10} lg={8}>
            <Title>Create new resource with template</Title>
          </Grid>
          <ResourceForm template={getTemplate(templateName)} />
        </Grid>
      </NotificationsWrapper>
    </Page>
  );
};

export default CreateResourcePage;
