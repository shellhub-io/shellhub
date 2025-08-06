/* eslint-disable */
// @ts-nocheck
function iota(start = 0) {
  let count = start;
  return new Proxy(
    {},
    {
      get(o, prop) {
        if (prop in o) return o[prop];

        count += 1;
        // eslint-disable-next-line no-return-assign,no-param-reassign
        return (o[prop] = count);
      },
    },
  );
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
  // Web Endpoint
  webEndpointCreate,
  webEndpointDelete,
  // Connector,
  connectorAdd,
  connectorEdit,
  connectorRemove,
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
  namespaceEdit,
  namespaceAddMember,
  namespaceEditMember,
  namespaceRemoveMember,
  namespaceEnableSessionRecord,
  namespaceRemove,
  NamespaceLeave,
  // Billing
  billingSubscribe,
  billingUnsubscribe,
  // Notification
  notificationView,
  apiKeyCreate,
  apiKeyDelete,
} = iota();

export const actions: { [key: string]: any } = {
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
  webEndpoint: {
    create: webEndpointCreate,
    delete: webEndpointDelete,
  },
  connector: {
   add: connectorAdd,
   edit: connectorEdit,
   remove: connectorRemove
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
    rename: namespaceEdit,
    addMember: namespaceAddMember,
    editMember: namespaceEditMember,
    removeMember: namespaceRemoveMember,
    enableSessionRecord: namespaceEnableSessionRecord,
    remove: namespaceRemove,
    leave: NamespaceLeave,
  },
  billing: {
    subscribe: billingSubscribe,
    unsubscribe: billingUnsubscribe,
  },
  notification: {
    view: notificationView,
  },
  apiKey: {
    create: apiKeyCreate,
    delete: apiKeyDelete,
  },
};

export const authorizer : { [key: string]: any } = {
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
      // Web Endpoint
      actions.webEndpoint.create,
      actions.webEndpoint.delete,
      // Connectors
      actions.connector.add,
      actions.connector.remove,
      actions.connector.edit,
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
      actions.apiKey.create,
      actions.apiKey.delete,
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
      // Web Endpoint
      actions.webEndpoint.create,
      actions.webEndpoint.delete,
      // Connectors
      actions.connector.add,
      actions.connector.remove,
      actions.connector.edit,
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
      actions.namespace.leave,
      actions.namespace.enableSessionRecord,
      // Billing
      // Notification
      actions.notification.view,
      actions.apiKey.create,
      actions.apiKey.delete,
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
      actions.namespace.leave,
      // Billing
      // Notification
      actions.notification.view,
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
      actions.namespace.leave,
      // Billing
      // Notification
    ],
  },
  role: {
    owner: "owner",
    administrator: "administrator",
    operator: "operator",
    observer: "observer",
  },
};
