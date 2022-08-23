import { Policy } from '../../../cluster-services/cluster_services.pb';
import Severity from '../Severity';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { darcula } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { usePolicyStyle } from '../PolicyStyles';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useFeatureFlags } from '@weaveworks/weave-gitops';

function HeaderSection({
  id,
  clusterName,
  tags,
  severity,
  category,
  targets,
  description,
  howToSolve,
  code,
  tenant,
}: Policy) {
  const classes = usePolicyStyle();
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  return (
    <>
      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <div className={classes.cardTitle}>Policy ID:</div>
        <span className={classes.body1}>{id}</span>
      </div>
      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <div className={classes.cardTitle}>Cluster Name:</div>
        <span className={classes.body1}>{clusterName}</span>
      </div>
      {flags.WEAVE_GITOPS_FEATURE_TENANCY === 'true' && tenant ? (
        <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
          <div className={classes.cardTitle}>Tenant:</div>
          <span className={classes.body1}>{tenant}</span>
        </div>
      ) : null}
      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <span className={classes.cardTitle}>Tags:</span>
        {!!tags && tags?.length > 0 ? (
          tags?.map(tag => (
            <span key={tag} className={classes.chip}>
              {tag}
            </span>
          ))
        ) : (
          <span className={classes.body1}>
            There is no tags for this policy
          </span>
        )}
      </div>
      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <div className={`${classes.cardTitle} ${classes.marginrightSmall}`}>
          Severity:
        </div>
        <Severity severity={severity || ''} />
      </div>
      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <div className={classes.cardTitle}>Category:</div>
        <span className={classes.body1}>{category}</span>
      </div>
      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <div className={classes.cardTitle}>Targeted K8s Kind:</div>
        {targets?.kinds?.map(kind => (
          <span key={kind} className={classes.chip}>
            {kind}
          </span>
        ))}
      </div>

      <hr />
      <div className={classes.sectionSeperator}>
        <div className={classes.cardTitle}>Description:</div>
        <ReactMarkdown
          children={description || ''}
          className={classes.editor}
        />
      </div>

      <div className={classes.sectionSeperator}>
        <div className={classes.cardTitle}>How to solve:</div>
        <ReactMarkdown
          children={howToSolve || ''}
          className={classes.editor}
          remarkPlugins={[remarkGfm]}
        />
      </div>

      <div className={classes.sectionSeperator}>
        <div className={classes.cardTitle}>Policy Code:</div>
        <div>
          <SyntaxHighlighter
            language="rego"
            style={darcula}
            wrapLongLines="pre-wrap"
            showLineNumbers={true}
            codeTagProps={{
              className: classes.code,
            }}
            customStyle={{
              height: '450px',
            }}
          >
            {code}
          </SyntaxHighlighter>
        </div>
      </div>
    </>
  );
}

export default HeaderSection;
