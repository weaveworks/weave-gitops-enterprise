import { useListCanaries } from '../../../contexts/ProgressiveDelivery';
import { useApplicationsCount } from '../../Applications/utils';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { SectionHeader } from '../../Layout/SectionHeader';
import { CanaryTable } from './Table';

const ProgressiveDelivery = () => {
  const applicationsCount = useApplicationsCount();

  const { error, data, isLoading } = useListCanaries();

  return (
    <>
      <SectionHeader
        className="count-header"
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: applicationsCount,
          },
          { label: 'Delivery', count: data?.canaries?.length },
        ]}
      />
      <ContentWrapper
        loading={isLoading}
        errors={data?.errors}
        errorMessage={error?.message}
      >
        {data?.canaries && <CanaryTable canaries={data.canaries} />}
      </ContentWrapper>
    </>
  );
};

export default ProgressiveDelivery;
