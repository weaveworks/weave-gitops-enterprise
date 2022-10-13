import React, { FC } from 'react';

const CostEstimation: FC<{
  isCostEstimationEnabled?: string;
}> = ({ isCostEstimationEnabled = 'false' }) => {
  return isCostEstimationEnabled === 'false' ? null : (
    <div className="costEstimation">
      <h2>Cost Estimation</h2>
    </div>
  );
};

export default CostEstimation;
