import http from '@/helpers/http';

const login = async (user) => http().post('/login', user);

export { login as default };
