import React, { FC } from 'react';
import { NameLink } from '../Shared';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import { Timestamp, SourceRefSourceKind, SourceDetail } from '@weaveworks/weave-gitops';

type Props = {
  name: string;
  namespace: string;
}

const WGApplicationsGitRepository: FC<Props> = ({ name, namespace }) => {
  const applicationsCount = useApplicationsCount();

  return (
    <PageTemplate documentTitle="WeGO · Git Repository">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: applicationsCount,
          },
        ]}
      />
      <ContentWrapper type="WG">
        <SourceDetail
          name={name}
          namespace={namespace}
          type={SourceRefSourceKind.GitRepository}
          info={(s: any) => [
            [
              "URL",
              <NameLink href={s.url}>
                {s.url}
              </NameLink>,
            ],
            ["Ref", s.reference?.branch],
            ["Last Updated", <Timestamp time={s.lastUpdatedAt} />],
            ["Cluster", "Default"],
            ["Namespace", s.namespace],
          ]}
        />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsGitRepository;
