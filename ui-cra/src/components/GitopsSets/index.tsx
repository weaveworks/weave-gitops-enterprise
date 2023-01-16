import { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import {
  AutomationsTable,
  LoadingPage,
  useListAutomations,
  theme,
} from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { makeStyles, createStyles } from '@material-ui/core';
import { useListConfigContext } from '../../contexts/ListConfig';
import useGitOpsSets from '../../hooks/gitopssets';

interface Size {
  size?: 'small';
}

const useStyles = makeStyles(() =>
  createStyles({
    externalIcon: {
      marginRight: theme.spacing.small,
    },
  }),
);

const GitopsSets: FC = () => {
  const { data: gitopssets, isLoading } = useGitOpsSets();
  const history = useHistory();
  const listConfigContext = useListConfigContext();
  const repoLink = listConfigContext?.repoLink || '';
  const classes = useStyles();

  return (
    <PageTemplate
      documentTitle="GitopsSets"
      path={[
        {
          label: 'GitopsSets',
        },
      ]}
    >
      <ContentWrapper errors={automations?.errors}>
        {isLoading ? (
          <LoadingPage />
        ) : (
          <AutomationsTable automations={gitopssets?.result} />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default GitopsSets;
