export default {
  /* eslint-disable no-param-reassign */
  install(Vue) {
    const copy = {
      command: 'Command',
      deviceSSHID: 'Device SSHID',
      tenantId: 'Tenant ID',
    };

    Vue.copy = copy;
    Vue.prototype.$copy = copy;
  },
  /* eslint-enable no-param-reassign */
};
