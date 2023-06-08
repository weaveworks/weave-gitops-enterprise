import { FC } from 'react';
import { Auth, UserGroupsTable } from '@weaveworks/weave-gitops';
import React from 'react';
import { NotificationsWrapper } from './Layout/NotificationsWrapper';
import { Page } from './Layout/App';

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
