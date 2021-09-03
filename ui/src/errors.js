export default {
  /* eslint-disable no-param-reassign */
  install(Vue) {
    const errors = {
      snackbar: {
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
        deviceTagCreate: 'creating tag',
        deviceTagList: 'list tag',
        deviceTagEdit: 'editing tag',
        deviceTagUpdate: 'updating tag',
        deviceTagDelete: 'deleting tag',
        deviceChoice: 'select device',
        sessionList: 'list session',
        sessionClose: 'close session',
        sessionPlay: 'play session',
        sessionDetails: 'details session',
        sessionRemoveRecord: 'deleting recorded session',
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
        addUser: 'creating account',
        loginFailed: 'login',
        logoutFailed: 'logout',
        recoverPassword: 'sending email',
        validationAccount: 'validation account',
        updatingAccount: 'updating account',
        subscription: 'subscription',
        updateSubscription: 'updating subscription',
        cancelSubscription: 'cancelling subscription',
      },
      form: {
        invalid: (field, param, extra) => {
          const types = {
            min: `Minimum characters for ${field} is ${extra}`,
            max: `Maximum characters for ${field} is ${extra}`,
            alphanum: `The ${field} is not in an alphanumeric standard`,
            ascii: `The ${field} is not in an ascii standard`,
            email: 'The format for email is invalid',
            nonStandardCharacters: `Your ${field} should be 3-30 characters long`,
            nameUsed: `This ${field} is already taken`,
            other: 'The format is invalid',
          };
          return types[param] || types.other;
        },
        conflict: (field) => `This ${field} already exists`,
      },
    };

    Vue.errors = errors;
    Vue.prototype.$errors = errors;
  },
  /* eslint-enable no-param-reassign */
};
