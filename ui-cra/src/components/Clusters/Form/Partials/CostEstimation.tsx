import React, { FC } from 'react';

const CostEstimation: FC<{
  isCostEstimationEnabled?: string;
}> = ({ isCostEstimationEnabled = 'false' }) => {
  return isCostEstimationEnabled === 'false' ? null : (
    <div className="costEstimation">Cost estimation</div>
  );
};

export default CostEstimation;
