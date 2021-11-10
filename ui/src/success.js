export default {
  /* eslint-disable no-param-reassign */
  install(Vue) {
    const success = {
      deviceRename: 'renaming device',
      deviceDelete: 'deleting device',
      deviceTagCreate: 'creating tag',
      deviceTagEdit: 'editing tag',
      deviceTagUpdate: 'updating tag',
      deviceTagDelete: 'deleting tag',
      deviceChooser: 'select device',
      sessionClose: 'closing session',
      sessionRemoveRecord: 'deleting recorded session',
      firewallRuleCreating: 'creating rule',
      firewallRuleEditing: 'editing rule',
      firewallRuleDeleting: 'deleting rule',
      publicKeyCreating: 'creating public key',
      publicKeyEditing: 'editing public key',
      publicKeyDeleting: 'deleting public key',
      privateKeyCreating: 'creating private key',
      privateKeyEditing: 'editing private key',
      privateKeyDeleting: 'deleting private key',
      profileData: 'updating data',
      profilePassword: 'updating password',
      namespaceCreating: 'creating namespace',
      namespaceNewMember: 'adding new member',
      namespaceEditMember: 'editing member',
      namespaceDelete: 'deleting namespace',
      namespaceEdit: 'editing namespace',
      namespaceRemoveUser: 'removing member',
      namespaceReload: 'reloading namespace',
      addUser: 'creating account',
      resendEmail: 'resend email',
      recoverPassword: 'sending email',
      validationAccount: 'validation account',
      updatingAccount: 'updating account',
      subscription: 'subscription',
      updateSubscription: 'updating subscription',
      cancelSubscription: 'cancelling subscription',
    };

    Vue.success = success;
    Vue.prototype.$success = success;
  },
  /* eslint-enable no-param-reassign */
};
