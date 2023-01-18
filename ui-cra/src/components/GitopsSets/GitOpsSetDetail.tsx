import { Dialog } from '@material-ui/core';
import {
  AppContext,
  Flex,
  InfoList,
  Metadata,
  ReconciliationGraph,
  RouterTab,
  SubRouterTabs,
} from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import { useRouteMatch } from 'react-router-dom';
import { GitOpsSet } from '../../api/gitopssets/types.pb';
