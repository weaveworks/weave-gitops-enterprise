import React, { FC, Ref, useEffect, useRef } from 'react';
import styled from 'styled-components';
import { contentCss } from './ContentWrapper';
import { NotificationData } from '../../contexts/Notifications';

const Footer = styled.div<{
  variant: NotificationData['variant'];
}>`
  ${contentCss}
  display: flex;
  flex-direction: column;
  align-items: center;
  background-color: ${props =>
    props.variant === 'danger' ? '#ffcccc' : '#C3EBDF'};
`;

export const FooterWrapper: FC<{ notification: NotificationData }> = ({
  notification,
}) => {
  const notificationRef: Ref<HTMLDivElement> = useRef(null);

  useEffect(() => {
    if (notification) {
      notificationRef?.current?.scrollIntoView({
        behavior: 'smooth',
        block: 'center',
      });
    }
  }, [notification]);

  return (
    <Footer ref={notificationRef} variant={notification.variant}>
      <div>{notification.message}</div>
    </Footer>
  );
};
