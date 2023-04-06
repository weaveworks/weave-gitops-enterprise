import { Alert } from '@material-ui/lab';
import { Flex } from '@weaveworks/weave-gitops';
import { useCheckCRDInstalled } from '../../contexts/ImageAutomation';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
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
    <PageTemplate
      documentTitle="Image Automation"
      path={[{ label: 'Image Automation' }]}
    >
      <ContentWrapper loading={isLoading}>
        {error && <Alert severity="error">{error.message}</Alert>}
        {!isCRDAvailable ? (
          <Flex center>
            <OnboardingImageAutomationMessage />
          </Flex>
        ) : (
          <ImageAutomation />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
}

export default ImageAutomationPage;
