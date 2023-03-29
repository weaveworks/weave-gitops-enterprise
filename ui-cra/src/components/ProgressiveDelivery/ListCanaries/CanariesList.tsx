import { useListCanaries } from '../../../contexts/ProgressiveDelivery';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { SectionHeader } from '../../Layout/SectionHeader';
import { CanaryTable } from './Table';

const ProgressiveDelivery = () => {
  const { data, isLoading } = useListCanaries();

  return (
    <>
      <SectionHeader
        className="count-header"
        path={[
          { label: 'Delivery' },
        ]}
      />
      <ContentWrapper loading={isLoading} errors={data?.errors}>
        {data?.canaries && <CanaryTable canaries={data.canaries} />}
      </ContentWrapper>
    </>
  );
};

export default ProgressiveDelivery;
