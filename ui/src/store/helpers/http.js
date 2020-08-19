import Axios from 'axios';
import router from '@/router/index';
import store from '..';

export default () => {
  const axios = Axios.create({
    baseURL: `${window.location.protocol}//${window.location.host}/api`,
    headers: {
      Authorization: `Bearer ${localStorage.getItem('token')}`,
    },
  });

  axios.interceptors.response.use(
    (response) => response,
    async (error) => {
      if (error.response.status === 401) {
        await store.dispatch('auth/logout');
        await router.push({ name: 'login' });
      }
      throw error;
    },
  );

  return axios;
};
