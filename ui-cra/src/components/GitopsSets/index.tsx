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

const useStyles = makeStyles(() =>
  createStyles({
    externalIcon: {
      marginRight: theme.spacing.small,
    },
  }),
);

const GitopsSets: FC = () => {
  const { data, isLoading } = useGitOpsSets();
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
      <ContentWrapper errors={data?.errors}>
        {isLoading ? (
          <LoadingPage />
        ) : (
          <AutomationsTable automations={data?.gitopssets} />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default GitopsSets;
