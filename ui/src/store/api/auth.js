import http from '@/store/helpers/http';

export const login = async (user) => http().post('/login', user);

export const info = async () => http().get('/auth/user');
