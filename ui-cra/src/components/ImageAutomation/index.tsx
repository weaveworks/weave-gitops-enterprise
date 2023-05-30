import { Alert } from '@material-ui/lab';
import { Flex, Page } from '@weaveworks/weave-gitops';
import { useCheckCRDInstalled } from '../../contexts/ImageAutomation';
import ImageAutomation from './ImageAutomation';
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
      {error && <Alert severity="error">{error.message}</Alert>}
      {!isCRDAvailable ? (
        <Flex center wide>
          <OnboardingImageAutomationMessage />
        </Flex>
      ) : (
        <ImageAutomation />
      )}
    </Page>
  );
}

export default ImageAutomationPage;
