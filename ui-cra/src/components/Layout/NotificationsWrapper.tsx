import { FC, useEffect } from 'react';
import styled from 'styled-components';
import { ListError } from '../../cluster-services/cluster_services.pb';
import { useVersionContext } from '../../contexts/ListConfig';
import useNotifications, {
  NotificationData,
} from '../../contexts/Notifications';
import {
  WarningIcon,
  WarningWrapper,
} from '../PolicyConfigs/PolicyConfigStyles';
import { AlertListErrors } from './AlertListErrors';
import Notifications from './Notifications';

const ENTITLEMENT_ERROR =
  'No entitlement was found for Weave GitOps Enterprise. Please contact sales@weave.works.';

const ENTITLEMENT_WARN =
  'Your entitlement for Weave GitOps Enterprise has expired, please contact sales@weave.works.';

export const Title = styled.h2`
  margin-top: 0px;
`;

interface Props {
  errors?: ListError[];
  notifications?: NotificationData[];
  warningMsg?: string;
}

export const NotificationsWrapper: FC<Props> = ({
  children,
  errors,
  warningMsg,
}) => {
  const versionResponse = useVersionContext();
  const { notifications, setNotifications } = useNotifications();

  useEffect(() => {
    if (versionResponse?.entitlement === ENTITLEMENT_WARN) {
      setNotifications([
        {
          message: {
            text: versionResponse.entitlement,
          },
          severity: 'warning',
        } as NotificationData,
      ]);
    }
  }, [versionResponse?.entitlement, setNotifications]);

  const topNotifications = notifications.filter(
    n => n.display !== 'bottom' && n.message.text !== ENTITLEMENT_ERROR,
  );
  const bottomNotifications = notifications.filter(n => n.display === 'bottom');

  return (
    <div style={{ width: '100%' }}>
      {errors && (
        <AlertListErrors
          errors={errors.filter(error => error.message !== ENTITLEMENT_ERROR)}
        />
      )}
      {!!warningMsg && (
        <WarningWrapper
          severity="warning"
          iconMapping={{
            warning: <WarningIcon />,
          }}
        >
          <span>{warningMsg}</span>
        </WarningWrapper>
      )}
      <Notifications notifications={topNotifications} />

      {children}

      {!!bottomNotifications.length && (
        <div style={{ paddingTop: '16px' }}>
          <Notifications notifications={bottomNotifications} />
        </div>
      )}
    </div>
  );
};
