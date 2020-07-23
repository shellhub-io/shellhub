import http from '@/store/helpers/http';

const getStats = async () => http().get('/stats');

export { getStats as default };
