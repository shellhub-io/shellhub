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

  axios.interceptors.request.use((config) => {
    store.dispatch('spinner/setStatus', true);
    return config;
  }, async (error) => {
    throw error;
  });

  axios.interceptors.response.use((response) => {
    store.dispatch('spinner/setStatus', false);
    return response;
  }, async (error) => {
    store.dispatch('spinner/setStatus', false);
    if (error.response.status === 401) {
      await store.dispatch('auth/logout');
      await router.push({ name: 'login' }).catch(() => {});
    }
    throw error;
  });

  return axios;
};
