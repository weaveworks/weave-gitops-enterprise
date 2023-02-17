import { FC } from 'react';
import { Auth, UserGroupsTable } from '@weaveworks/weave-gitops';
import { ContentWrapper } from './Layout/ContentWrapper';
import { PageTemplate } from './Layout/PageTemplate';
import React from 'react';

const WGUserInfo: FC = () => {
  const { userInfo, error } = React.useContext(Auth);

  return (
    <PageTemplate
      documentTitle="User Info"
      path={[
        {
          label: 'User Info',
        },
      ]}
    >
      <ContentWrapper errors={error ? [{ message: error?.statusText }] : []}>
        <UserGroupsTable rows={userInfo?.groups} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGUserInfo;
