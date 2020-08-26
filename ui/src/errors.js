export default {
  /* eslint-disable no-param-reassign */
  install(Vue) {
    const errors = {
      dashboard: 'dashboard',
      notificationList: 'list notification',
      deviceList: 'list device',
      deviceAcess: 'accessing device',
      deviceDelete: 'deleting device',
      deviceListPending: 'list of pending device',
      devicePending: 'pending device',
      deviceAccepting: 'accepting device',
      deviceRejecting: 'rejecting device',
      deviceListRejected: 'list of rejected device',
      deviceDetails: 'details device',
      deviceRename: 'rename device',
      sessionList: 'list session',
      sessionClose: 'close session',
      sessionPlay: 'play session',
      sessionDetails: 'details session',
      firewallRuleList: 'list of firewall rule',
      firewallRuleCreating: 'creating rule',
      firewallRuleEditing: 'editing rule',
      firewallRuleDeleting: 'deleting rule',
    };

    Vue.errors = errors;
    Vue.prototype.$errors = errors;
  },
  /* eslint-enable no-param-reassign */
};
