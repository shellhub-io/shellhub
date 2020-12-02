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
      publicKeyList: 'list of public key',
      publicKeyCreating: 'creating public key',
      publicKeyEditing: 'editing public key',
      publicKeyDeleting: 'deleting public key',
      namespaceList: 'namespace list',
      namespaceCreating: 'creating namespace',
      namespaceNewMember: 'adding new member',
      namespaceDelete: 'deleting namespace',
      namespaceSwitch: 'switching namespace',
      namespaceEdit: 'editing namespace',
      namespaceRemoveUser: 'removing user',
      namespaceLoad: 'loading namespace',
      privateKeyList: 'list of private key',
      privateKeyCreating: 'creating private key',
      privateKeyEditing: 'editing private key',
      privateKeyDeleting: 'deleting private key',
    };

    Vue.errors = errors;
    Vue.prototype.$errors = errors;
  },
  /* eslint-enable no-param-reassign */
};
