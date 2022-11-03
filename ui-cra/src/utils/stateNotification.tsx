import React from 'react';
import { Link } from '@weaveworks/weave-gitops';
import { NotificationData } from '../types/custom';

export const stateNotification = (notification: NotificationData) => {
  if (notification?.message?.text) {
    if (notification?.message?.text?.includes('PR created successfully')) {
      const href = notification?.message?.text.split('::')[1];
      return {
        message: {
          component: (
            <Link href={href} newTab>
              PR created successfully.
            </Link>
          ),
        },
        severity: 'success',
      } as NotificationData;
    }
    return {
      message: {
        text: notification?.message?.text,
      },
      severity: notification?.severity,
    };
  } else
    return {
      message: {
        component: notification?.message?.component,
      },
      severity: notification?.severity,
    };
};
