import { useListCanaries } from '../../../contexts/ProgressiveDelivery';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { SectionHeader } from '../../Layout/SectionHeader';
import { CanaryTable } from './Table';

const ProgressiveDelivery = () => {
  const { error, data, isLoading } = useListCanaries();

  return (
    <>
      <SectionHeader
        className="count-header"
        path={[
          {
            label: 'Applications',
            url: '/applications',
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
