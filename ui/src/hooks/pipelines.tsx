import { useQuery } from 'react-query';
import {
  ApprovePromotionRequest,
  ApprovePromotionResponse,
  Pipelines,
} from '../api/pipelines/pipelines.pb';
import { toast } from 'react-toastify';

export const useApprove = (req: ApprovePromotionRequest) => {
  const { data, isLoading, error, refetch } = useQuery<
    ApprovePromotionResponse,
    Error
  >('approve', () => Pipelines.ApprovePromotion(req), {
    enabled: false,
    onError: (error: Error) =>
      toast['error'](error.message || 'Error promoting pipeline'),
  });
  return { data, isLoading, error, refetch };
};
