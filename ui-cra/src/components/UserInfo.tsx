import { FC } from 'react';
import { Auth, Page, UserGroupsTable } from '@weaveworks/weave-gitops';
import { ContentWrapper } from './Layout/ContentWrapper';
import React from 'react';

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
      <ContentWrapper errors={error ? [{ message: error?.statusText }] : []}>
        <UserGroupsTable rows={userInfo?.groups} />
      </ContentWrapper>
    </Page>
  );
};

export default WGUserInfo;
