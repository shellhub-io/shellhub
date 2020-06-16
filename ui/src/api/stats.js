import http from '@/helpers/http';

const getStats = async () => http().get('/stats');

export { getStats as default };
