import Axios from 'axios';
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
    (error) => {
      if (error.response.status === 401) {
        store.dispatch('auth/logout');
      }
      throw error;
    },
  );

  return axios;
};
