export default {
  /* eslint-disable no-param-reassign */
  install(Vue) {
    const success = {
      deviceRename: 'renaming device',
      deviceDelete: 'deleting device',
      sessionClose: 'closing session',
      firewallRuleCreating: 'creating rule',
      firewallRuleEditing: 'editing rule',
      firewallRuleDeleting: 'deleting rule',
      profileData: 'updating data',
      profilePassword: 'updating password',
      NamespaceCreating: 'creating namespace',
    };

    Vue.success = success;
    Vue.prototype.$success = success;
  },
  /* eslint-enable no-param-reassign */
};
