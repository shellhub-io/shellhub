import http from '@/store/helpers/http';

export default async () => http().get('/stats');
