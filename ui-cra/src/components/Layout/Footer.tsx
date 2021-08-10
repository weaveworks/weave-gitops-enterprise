import React, { FC, Ref, useEffect, useRef, useState } from 'react';
import styled from 'styled-components';
import { NotificationData } from '../../contexts/Notifications';
import IconButton from '@material-ui/core/IconButton';
import { Close } from '@material-ui/icons';
import useNotifications from '../../contexts/Notifications';
import { contentCss } from './ContentWrapper';

const Footer = styled.div<{
  variant: NotificationData['variant'];
}>`
  ${contentCss}
  display: flex;
  flex-direction: column;
  background-color: ${props =>
    props.variant === 'danger' ? '#ffcccc' : '#C3EBDF'};
`;

const CloseIconBox = styled.div`
  display: flex;
  justify-content: flex-end;
  align-items: center;
  height: 20px;
`;

const NotificationBox = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
`;

export const FooterWrapper: FC<{ notification: NotificationData }> = ({
  notification,
}) => {
  const notificationRef: Ref<HTMLDivElement> = useRef(null);
  const [open, setOpen] = useState<boolean>(true);
  const { setNotification } = useNotifications();

  useEffect(() => {
    setOpen(true);
    if (notification) {
      notificationRef?.current?.scrollIntoView({
        behavior: 'smooth',
        block: 'center',
      });
    }
  }, [notification]);

  const onClose = () => {
    setOpen(false);
    setNotification(null);
  };

  return open ? (
    <Footer ref={notificationRef} variant={notification.variant}>
      <CloseIconBox>
        <IconButton onClick={onClose}>
          <Close />
        </IconButton>
      </CloseIconBox>
      <NotificationBox>{notification.message}</NotificationBox>
    </Footer>
  ) : null;
};
