import { Policy } from '../../../cluster-services/cluster_services.pb';
import Severity from '../Severity';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { darcula } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { usePolicyStyle } from '../PolicyStyles';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { generateRowHeaders } from '../../ProgressiveDelivery/SharedComponent/CanaryRowHeader';

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
}: Policy) {
  const classes = usePolicyStyle();

  const defaultHeaders = [
    {
      rowkey: 'Policy ID',
      value: id,
    },
    {
      rowkey: 'Cluster Name',
      value: clusterName,
    },
    {
      rowkey: 'Tags',
      children: (
        <>
          {!!tags && tags?.length > 0 ? (
            tags?.map(tag => (
              <span key={tag} className={classes.chip}>
                {tag}
              </span>
            ))
          ) : (
            <span>There is no tags for this policy</span>
          )}
        </>
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
        <>
          {targets?.kinds?.map(kind => (
            <span key={kind} className={classes.chip}>
              {kind}
            </span>
          ))}
        </>
      ),
    },
  ];

  return (
    <>
      {generateRowHeaders(defaultHeaders)}

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
