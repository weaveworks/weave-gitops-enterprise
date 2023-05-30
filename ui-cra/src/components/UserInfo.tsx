import { FC } from 'react';
import { Auth, Page, UserGroupsTable } from '@weaveworks/weave-gitops';
import React from 'react';

const WGUserInfo: FC = () => {
  const { userInfo, error } = React.useContext(Auth);

  return (
    <Page
      error={error ? [{ message: error?.statusText }] : []}
      path={[
        {
          label: 'User Info',
        },
      ]}
    >
      <UserGroupsTable rows={userInfo?.groups} />
    </Page>
  );
};

export default WGUserInfo;
