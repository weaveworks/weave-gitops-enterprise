import { Auth, UserGroupsTable } from '@weaveworks/weave-gitops';
import React, { FC } from 'react';
import { Page } from './Layout/App';
import { NotificationsWrapper } from './Layout/NotificationsWrapper';

const WGUserInfo: FC = () => {
  const { userInfo, error } = React.useContext(Auth);

  return (
    <Page
      path={[
        {
          label: 'User Info',
        },
      ]}
    >
      <NotificationsWrapper
        errors={error ? [{ message: error?.statusText }] : []}
      >
        <UserGroupsTable rows={userInfo?.groups} />
      </NotificationsWrapper>
    </Page>
  );
};

export default WGUserInfo;
