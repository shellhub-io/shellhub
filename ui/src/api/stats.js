import http from '@/helpers/http';

export const
  getStats = async () => {
    return http().get('/stats');
  };