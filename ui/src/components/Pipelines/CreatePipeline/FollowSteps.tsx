import { Text } from '@weaveworks/weave-gitops';
import React from 'react';

export const FollowSteps = () => {
  return (
    <div>
      <Text semiBold>Follow the steps:</Text>
      <ol>
        <li>Click the "Create Environment" button at right</li>
        <li>Fill out the fields that appear in the left column</li>
        <li>When you're done, click the "Apply" button</li>
        <li>
          Click the "Add Target" button to add a target—filling out the fields
          and clicking "Apply" when you're finished-. Repeat this step as many
          times as you have targets.
        </li>
        <li>
          To add another environment, go back to Step One of this list and
          repeat.
        </li>
        <li>
          Review the information under "GitOps: Review details and create."
        </li>
        <li>Click the "Apply" button at the bottom of the screen.</li>
        <li>
          To Edit your Pipeline, find the environment or target at right and
          click the gear icon. Edit your information in the fields.
        </li>
      </ol>
    </div>
  );
};
