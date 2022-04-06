import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { CallbackStateContextProvider } from '@weaveworks/weave-gitops';
import { ContentWrapper, Title } from '../../Layout/ContentWrapper';

import { PolicyService } from './../PolicyService';
import { useState } from 'react';
import LoadingError from '../../LoadingError';
import { useParams } from 'react-router-dom';
import { GetPolicyResponse, Policy } from '../../../capi-server/capi_server.pb';
import { createStyles, makeStyles } from '@material-ui/styles';

const useStyles = makeStyles(() =>
  createStyles({
    cardTitle: {
      fontWeight: 700,
      fontSize: '14px',
      color: '#737373',
      marginBottom: '12px',
    },
    body1: {
      fontWeight: 400,
      fontSize: '14px',
      color: '#1A1A1A',
      marginLeft: '8px',
    },
    chip: {
      background: 'rgba(10, 57, 64, 0.06)',
      borderRadius: '2px',
      padding: '2px 8px',
      marginLeft: '8px',
      fontWeight: 400,
      fontSize: '11px',
    },
  }),
);

// Move to separate file => PolicyDetailsHeaderSection
const HeaderSection = ({ id, tags, severity, category, targets }: Policy) => {
  const classes = useStyles();
  return (
    <>
      <div>
        <span className={classes.cardTitle}>Policy ID:</span>
        <span className={classes.body1}>{id}</span>
      </div>
      <div>
        <span className={classes.cardTitle}>Tags:</span>
        {tags?.map(tag => (
          <span key={tag} className={classes.chip}>
            {tag}
          </span>
        ))}
      </div>
      <div>
        <span className={classes.cardTitle}>Severity:</span>
        <span className={classes.body1}>{severity}</span>
      </div>
      <div>
        <span className={classes.cardTitle}>Category:</span>
        <span className={classes.body1}>{category}</span>
      </div>
      <div>
        <span className={classes.cardTitle}>Targeted K8s Kind:</span>
        {targets?.kinds?.map(kind => (
          <span key={kind} className={classes.chip}>
            {kind}
          </span>
        ))}
      </div>
    </>
  );
};

const PolicyDetails = () => {
  const { id } = useParams<{ id: string }>();
  const [name, setName] = useState('');

  const fetchPoliciesAPI = () =>
    PolicyService.getPolicyById(id).then((res: GetPolicyResponse) => {
      res.policy?.name && setName(res.policy.name);
      return res;
    });

  const [fetchPolicyById] = useState(() => fetchPoliciesAPI);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Policies">
        <CallbackStateContextProvider>
          <SectionHeader
            className="count-header"
            path={[
              { label: 'Policies', url: '/policies' },
              { label: name, url: 'policy-details' },
            ]}
          />
          <ContentWrapper>
            <Title>{name}</Title>
            <LoadingError fetchFn={fetchPolicyById}>
              {({ value: { policy } }: { value: GetPolicyResponse }) => (
                <>
                  <HeaderSection
                    id={policy?.id}
                    tags={policy?.tags}
                    severity={policy?.severity}
                    category={policy?.category}
                    targets={policy?.targets}
                  />
                </>
              )}
            </LoadingError>
          </ContentWrapper>
        </CallbackStateContextProvider>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PolicyDetails;
