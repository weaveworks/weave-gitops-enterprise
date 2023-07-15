import { useLocation } from 'react-router-dom';
import useTemplates from '../../../hooks/templates';
import { NotificationsWrapper, Title } from '../../Layout/NotificationsWrapper';
import ResourceForm from '../Form';
import { Page } from '../../Layout/App';

const CreateResourcePage = () => {
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
        <Title>Create new resource with template</Title>
        <ResourceForm template={getTemplate(templateName)} />
      </NotificationsWrapper>
    </Page>
  );
};

export default CreateResourcePage;
