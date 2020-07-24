import http from '@/store/helpers/http';

const login = async (user) => http().post('/login', user);
const info = async () => http().get('/auth/user');

export { login, info };
