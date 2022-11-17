import { Policy } from '../../../cluster-services/cluster_services.pb';
import Severity from '../Severity';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { darcula } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { usePolicyStyle } from '../PolicyStyles';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';
import { useFeatureFlags } from '@weaveworks/weave-gitops';
import Mode from '../Mode';

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
  modes,
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
        <div id="policy-details-header-tags">
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
      rowkey: 'Mode',
      children: modes?.length
        ? modes.map((mode: string, index: number) => (
            <Mode key={index} modeName={mode} showName={true} />
          ))
        : '',
    },
    {
      rowkey: 'Targeted K8s Kind',
      children: (
        <div id="policy-details-header-kinds">
          {targets?.kinds?.length ? (
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
