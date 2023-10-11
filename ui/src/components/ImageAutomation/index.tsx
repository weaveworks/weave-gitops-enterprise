import { Alert } from '@material-ui/lab';
import { Flex, ImageAutomation, useCheckCRDInstalled } from '@weaveworks/weave-gitops';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import OnboardingImageAutomationMessage from './OnboardingMessage';
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
