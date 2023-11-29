import { Flex, Text } from '@weaveworks/weave-gitops';
import React from 'react';
import styled from 'styled-components';

const steps = [
  'Click the "Create Environment" button at right',
  'Fill out the fields that appear in the left column',
  `When you're done, click the "Apply" button`,
  `Click the "Add Target" button to add a target—filling out the fields
and clicking "Apply" when you're finished-. Repeat this step as many
times as you have targets.`,
  'To add another environment, go back to Step One of this list and repeat.',
  'Review the information under "GitOps: Review details and create."',
  'Click the "Apply" button at the bottom of the screen.',
  'To Edit your Pipeline, find the environment or target at right and click the gear icon. Edit your information in the fields.',
];

const StepsContainer = styled(Flex)`
  padding: 8px;
`;
export const FollowSteps = () => {
  return (
    <div>
      <Text semiBold>Follow the steps:</Text>
      <StepsContainer column gap="8">
        {steps.map((s, i) => (
          <Text color="grayToPrimary" key={s}>
            {i + 1}. {s}
          </Text>
        ))}
      </StepsContainer>
    </div>
  );
};
