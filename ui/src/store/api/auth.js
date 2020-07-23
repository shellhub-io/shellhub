import http from '@/store/helpers/http';

const login = async (user) => http().post('/login', user);

export { login as default };
