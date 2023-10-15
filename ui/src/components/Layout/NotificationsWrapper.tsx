import { NotificationsWrapperOSS } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import styled from 'styled-components';
import { ListError } from '../../cluster-services/cluster_services.pb';
import { useVersionContext } from '../../contexts/ListConfig';
import { NotificationData } from '../../contexts/Notifications';
import useNotifications from '../../contexts/Notifications';

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

  return (
    <NotificationsWrapperOSS
      children={children}
      errors={errors}
      warningMsg={warningMsg}
      notifications={notifications}
      setNotifications={setNotifications}
      versionEntitlement={
        versionResponse?.entitlement === ENTITLEMENT_WARN
          ? versionResponse?.entitlement
          : ''
      }
    />
  );
};
