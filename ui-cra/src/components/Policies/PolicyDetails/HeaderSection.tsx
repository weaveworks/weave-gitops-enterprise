import { Policy } from '../../../cluster-services/cluster_services.pb';
import Severity from '../Severity';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { darcula } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { usePolicyStyle } from '../PolicyStyles';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import {
  generateRowHeaders,
  SectionRowHeader,
} from '../../ProgressiveDelivery/SharedComponent/CanaryRowHeader';
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

  const defaultHeaders: Array<SectionRowHeader> = [
    {
      rowkey: 'Policy ID',
      value: id,
    },
    {
      rowkey: 'Cluster Name',
      value: clusterName,
    },
    {
      rowkey: 'Tenant',
      value: tenant,
      hidden: flags.WEAVE_GITOPS_FEATURE_TENANCY !== 'true',
    },
    {
      rowkey: 'Tags',
      children: (
        <div id="tags">
          {!!tags && tags?.length > 0 ? (
            tags?.map(tag => (
              <span key={tag} className={classes.chip}>
                {tag}
              </span>
            ))
          ) : (
            <span>There is no tags for this policy</span>
          )}
        </div>
      ),
    },
    {
      rowkey: 'Severity',
      children: <Severity severity={severity || ''} />,
    },
    {
      rowkey: 'Category',
      value: category,
    },
    {
      rowkey: 'Targeted K8s Kind',
      children: (
        <div id="kinds">
          {!!targets?.kinds && targets.kinds.length > 0 ? (
            targets?.kinds?.map(kind => (
              <span key={kind} className={classes.chip}>
                {kind}
              </span>
            ))
          ) : (
            <span>There is no kinds for this policy</span>
          )}
        </div>
      ),
    },
  ];

  return (
    <>
      {generateRowHeaders(defaultHeaders)}

      <div className={classes.sectionSeperator} data-testid="description">
        <div className={classes.cardTitle}>Description:</div>
        <ReactMarkdown
          children={description || ''}
          className={classes.editor}
        />
      </div>

      <div className={classes.sectionSeperator} data-testid="howToSolve">
        <div className={classes.cardTitle}>How to solve:</div>
        <ReactMarkdown
          children={howToSolve || ''}
          className={classes.editor}
          remarkPlugins={[remarkGfm]}
        />
      </div>

      <div className={classes.sectionSeperator} data-testid="policyCode">
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
