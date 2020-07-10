export default {
  /* eslint-disable no-param-reassign */
  install(Vue) {
    const env = {
      isHosted: window.env.HOSTED === 'true',
    };

    Vue.env = env;
    Vue.prototype.$env = env;
  },
  /* eslint-enable no-param-reassign */
};
