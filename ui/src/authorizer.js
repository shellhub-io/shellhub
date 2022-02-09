function iota(start = 0) {
  let count = start;
  return new Proxy({}, {
    get(o, prop) {
      if (prop in o) return o[prop];

      count += 1;
      // eslint-disable-next-line no-return-assign,no-param-reassign
      return o[prop] = count;
    },
  });
}

const {
  // Device
  deviceAdd,
  deviceAccept,
  deviceReject,
  deviceRemove,
  deviceConnect,
  deviceRename,
  deviceChooser,
  deviceDetails,
  // Tag
  tagEdit,
  tagRemove,
  tagDeviceCreate,
  tagDeviceUpdate,
  // Session
  sessionPlay,
  sessionClose,
  sessionDetails,
  sessionRemoveRecord,
  // Firewall
  firewallCreate,
  firewallEdit,
  firewallRemove,
  // Public Key
  publicKeyCreate,
  publicKeyEdit,
  publicKeyRemove,
  // Namespace
  namespaceCreate,
  namespaceRename,
  namespaceAddMember,
  namespaceEditMember,
  namespaceRemoveMember,
  namespaceEnableSessionRecord,
  namespaceRemove,
  // Billing
  billingSubscribe,
  billingUnsubscribe,
  // Notification
  notificaitonView,
} = iota();

export const actions = {
  device: {
    add: deviceAdd,
    accept: deviceAccept,
    reject: deviceReject,
    remove: deviceRemove,
    connect: deviceConnect,
    rename: deviceRename,
    chooser: deviceChooser,
    details: deviceDetails,
  },
  tag: {
    edit: tagEdit,
    remove: tagRemove,
    deviceCreate: tagDeviceCreate,
    deviceUpdate: tagDeviceUpdate,
  },
  session: {
    play: sessionPlay,
    close: sessionClose,
    details: sessionDetails,
    removeRecord: sessionRemoveRecord,
  },
  firewall: {
    create: firewallCreate,
    edit: firewallEdit,
    remove: firewallRemove,
  },
  publicKey: {
    create: publicKeyCreate,
    edit: publicKeyEdit,
    remove: publicKeyRemove,
  },
  namespace: {
    create: namespaceCreate,
    rename: namespaceRename,
    addMember: namespaceAddMember,
    editMember: namespaceEditMember,
    removeMember: namespaceRemoveMember,
    enableSessionRecord: namespaceEnableSessionRecord,
    remove: namespaceRemove,
  },
  billing: {
    subscribe: billingSubscribe,
    unsubscribe: billingUnsubscribe,
  },
  notification: {
    view: notificaitonView,
  },
};

export const authorizer = {
  permissions: {
    owner: [
      // Device
      actions.device.add,
      actions.device.accept,
      actions.device.reject,
      actions.device.remove,
      actions.device.connect,
      actions.device.rename,
      actions.device.chooser,
      actions.device.details,
      // Tag
      actions.tag.edit,
      actions.tag.remove,
      actions.tag.deviceCreate,
      actions.tag.deviceUpdate,
      // Session
      actions.session.play,
      actions.session.close,
      actions.session.details,
      actions.session.removeRecord,
      // Firewall
      actions.firewall.create,
      actions.firewall.edit,
      actions.firewall.remove,
      // Public Key
      actions.publicKey.create,
      actions.publicKey.edit,
      actions.publicKey.remove,
      // Namespace
      actions.namespace.create,
      actions.namespace.rename,
      actions.namespace.addMember,
      actions.namespace.editMember,
      actions.namespace.removeMember,
      actions.namespace.enableSessionRecord,
      actions.namespace.remove,
      // Billing
      actions.billing.subscribe,
      actions.billing.unsubscribe,
      // Notification
      actions.notification.view,
    ],
    administrator: [
      // Device
      actions.device.add,
      actions.device.accept,
      actions.device.reject,
      actions.device.remove,
      actions.device.connect,
      actions.device.rename,
      actions.device.details,
      // Tag
      actions.tag.edit,
      actions.tag.remove,
      actions.tag.deviceCreate,
      actions.tag.deviceUpdate,
      // Session
      actions.session.play,
      actions.session.close,
      actions.session.details,
      actions.session.removeRecord,
      // Firewall
      actions.firewall.create,
      actions.firewall.edit,
      actions.firewall.remove,
      // Public Key
      actions.publicKey.create,
      actions.publicKey.edit,
      actions.publicKey.remove,
      // Namespace
      actions.namespace.create,
      actions.namespace.rename,
      actions.namespace.addMember,
      actions.namespace.editMember,
      actions.namespace.removeMember,
      actions.namespace.enableSessionRecord,
      // Billing
      // Notification
      actions.notification.view,
    ],
    operator: [
      // Device
      actions.device.add,
      actions.device.accept,
      actions.device.reject,
      actions.device.connect,
      actions.device.rename,
      actions.device.details,
      // Tag
      actions.tag.edit,
      actions.tag.remove,
      actions.tag.deviceCreate,
      actions.tag.deviceUpdate,
      // Session
      actions.session.details,
      // Firewall
      // Public Key
      // Namespace
      actions.namespace.create,
      // Billing
      // Notification
    ],
    observer: [
      // Device
      actions.device.connect,
      actions.device.details,
      // Tag
      // Session
      actions.session.details,
      // Firewall
      // Public Key
      // Namespace
      actions.namespace.create,
      // Billing
      // Notification
    ],
  },
  role: {
    owner: 'owner',
    administrator: 'administrator',
    operator: 'operator',
    observer: 'observer',
  },
};

export const authorizerVue = {
  /* eslint-disable no-param-reassign */
  install(Vue) {
    Vue.authorizer = authorizer;
    Vue.actions = actions;
    Vue.prototype.$authorizer = authorizer;
    Vue.prototype.$actions = actions;
  },
  /* eslint-enable no-param-reassign */
};
