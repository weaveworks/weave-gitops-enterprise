import { Alert } from '@material-ui/lab';
import { Flex } from '@weaveworks/weave-gitops';
import { useCheckCRDInstalled } from '../../contexts/ImageAutomation';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import ImageAutomation from './ImageAutomation';
import OnboardingImageAutomationMessage from './OnboardingMessage';
import { Page } from '../Layout/App';
const crdName = 'imageupdateautomations.image.toolkit.fluxcd.io';

function ImageAutomationPage() {
  const {
    data: isCRDAvailable,
    isLoading,
    error,
  } = useCheckCRDInstalled(crdName);
  return (
    <Page loading={isLoading} path={[{ label: 'Image Automation' }]}>
      <NotificationsWrapper>
        {error && <Alert severity="error">{error.message}</Alert>}
        {!isCRDAvailable ? (
          <Flex center wide>
            <OnboardingImageAutomationMessage />
          </Flex>
        ) : (
          <ImageAutomation />
        )}
      </NotificationsWrapper>
    </Page>
  );
}

export default ImageAutomationPage;
