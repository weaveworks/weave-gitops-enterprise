import { Tooltip } from '@material-ui/core';
import { Flex, Icon, IconType, Link, Text } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

const EllipsesText = styled(Text)<{ maxWidth?: string }>`
  max-width: ${prop => prop.maxWidth || '400px'};
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;
export interface Breadcrumb {
  label: string;
  url?: string;
}
interface Props {
  path: Breadcrumb[];
}
export const Breadcrumbs = ({ path = [] }: Props) => {
  return (
    <Flex align>
      {path.map(({ label, url }) => {
        return (
          <Flex align key={label}>
            {url ? (
              <>
                <Link
                  data-testid={`link-${label}`}
                  to={url}
                  textProps={{ bold: true, size: 'large', color: 'neutral40' }}
                >
                  {label}
                </Link>
                <Icon
                  type={IconType.NavigateNextIcon}
                  size="large"
                  color="neutral40"
                />
              </>
            ) : (
              <Tooltip title={label} placement="bottom">
                <EllipsesText
                  size="large"
                  color="neutral40"
                  className="ellipsis"
                  data-testid={`text-${label}`}
                >
                  {label}
                </EllipsesText>
              </Tooltip>
            )}
          </Flex>
        );
      })}
    </Flex>
  );
};

export default styled(Breadcrumbs).attrs({ className: Breadcrumbs.name })``;
