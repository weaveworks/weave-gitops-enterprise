import { Box } from '@material-ui/core';
import * as React from 'react';
import styled from 'styled-components';

type Props = {
  className?: string;
  children?: React.ReactNode;
  open?: boolean;
  onClose?: () => void;
};

const SlideContainer = styled.div`
  position: relative;
  height: 100%;
  transition-property: width, left;
  transition-duration: 0.5s;
  transition-timing-function: linear;
  overflow: hidden;
  width: 0px;
  flex: 2;
  left: 24px;

  &.open {
    left: 0;
    width: 360px;
  }
`;

const SlideContent = styled.div`
  height: 100%;
  width: 100%;
  border-left: 2px solid ${props => props.theme.colors.neutral20};
  padding-left: ${props => props.theme.spacing.large};
`;

function FilterDrawer({ className, children, open, onClose }: Props) {
  return (
    <div className={className}>
      <SlideContainer className={open ? 'open' : ''}>
        <SlideContent>
          <Box p={1}>
            <div>
              <h2>Filters</h2>
            </div>
            <div>{children}</div>
          </Box>
        </SlideContent>
      </SlideContainer>
    </div>
  );
}

export default styled(FilterDrawer).attrs({ className: FilterDrawer.name })`
  overflow: hidden;
  white-space: nowrap;
`;
