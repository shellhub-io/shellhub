import Vuex from 'vuex';
import { shallowMount, createLocalVue, config } from '@vue/test-utils';
import SettingNamespace from '@/components/setting/SettingNamespace';

config.mocks = {
  $env: {
    isEnterprise: true,
  },
};

describe('SettingNamespace', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapperOwner;
  let wrapperNotOwner;
  let wrapperOwnerOpen;

  const idOwner = '6';
  const idNotOwner = '10';
  const owner = true;
  const textRole = ['Owner', 'Member', 'Member'];

  const countDevicesHasNamespacePercent = {
    maxDevices: 0,
    percent: 0,
  };

  const namespace = {
    name: 'namespace3',
    members: [{ id: '6', name: 'user6' }, { id: '7', name: 'user7' }, { id: '8', name: 'user8' }],
    owner: '6',
    tenant_id: 'e359bf484715',
    devices_count: 1,
    max_devices: 3,
  };

  const openNamespace = {
    name: 'namespace3',
    members: [{ id: '6', name: 'user6' }, { id: '7', name: 'user7' }, { id: '8', name: 'user8' }],
    owner: '6',
    tenant_id: 'e359bf484715',
    devices_count: 1,
    max_devices: -1,
  };

  const storeNotOwner = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      id: idNotOwner,
      owner: !owner,
      tenant: '',
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
      'auth/id': (state) => state.id,
      'namespaces/owner': (state) => state.owner,
      'auth/tenant': (state) => state.tenant,
    },
    actions: {
      'namespaces/put': () => {
      },
      'namespaces/get': () => {
      },
      'namespaces/removeUser': () => {
      },
      'snackbar/showSnackbarSuccessAction': () => {
      },
      'snackbar/showSnackbarErrorAction': () => {
      },
    },
  });

  const storeOwner = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      id: idOwner,
      owner,
      tenant: openNamespace.tenant_id,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
      'auth/id': (state) => state.id,
      'namespaces/owner': (state) => state.owner,
      'auth/tenant': (state) => state.tenant,
    },
    actions: {
      'namespaces/put': () => {
      },
      'namespaces/get': () => {
      },
      'namespaces/removeUser': () => {
      },
      'snackbar/showSnackbarSuccessAction': () => {
      },
      'snackbar/showSnackbarErrorAction': () => {
      },
    },
  });

  const storeOwnerOpen = new Vuex.Store({
    namespaced: true,
    state: {
      namespace: openNamespace,
      id: idOwner,
      owner,
      tenant: openNamespace.tenant_id,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
      'auth/id': (state) => state.id,
      'namespaces/owner': (state) => state.owner,
      'auth/tenant': (state) => state.tenant,
    },
    actions: {
      'namespaces/put': () => {
      },
      'namespaces/get': () => {
      },
      'namespaces/removeUser': () => {
      },
      'snackbar/showSnackbarSuccessAction': () => {
      },
      'snackbar/showSnackbarErrorAction': () => {
      },
    },
  });

  beforeEach(() => {
    jest.spyOn(Storage.prototype, 'getItem').mockReturnValue('e359bf484715');

    wrapperOwner = shallowMount(SettingNamespace, {
      store: storeOwner,
      localVue,
      stubs: ['fragment'],
      mocks: ['$env'],
    });
    wrapperNotOwner = shallowMount(SettingNamespace, {
      localVue,
      store: storeNotOwner,
      stubs: ['fragment'],
      mocks: ['$env'],
    });
    wrapperOwnerOpen = shallowMount(SettingNamespace, {
      localVue,
      store: storeOwnerOpen,
      stubs: ['fragment'],
      mocks: {
        $env: {
          isEnterprise: false,
        },
      },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapperOwner).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapperOwner.html()).toMatchSnapshot();
  });
  it('Renders the template with data - hosted version tests', () => {
    expect(wrapperOwner.find('[data-test=tenant]').text()).toEqual(namespace.tenant_id);
  });
  it('Process data in the computed - hosted version tests', () => {
    expect(wrapperOwner.vm.namespace).toEqual(namespace);
    expect(wrapperOwner.vm.tenant).toEqual(namespace.tenant_id);
    expect(wrapperOwner.vm.isEnterpriseOwner).toEqual(true);
    expect(wrapperOwner.vm.isEnterprise).toEqual(true);
  });
  it('Process data in methods - hosted version tests', () => {
    expect(wrapperOwner.vm.hasTenant()).toEqual(true);
  });
  it('Loads name when component is created - hosted version tests', () => {
    wrapperOwner.vm.$nextTick(() => {
      expect(wrapperOwner.vm.name).toBe(namespace.name);
    });
  });
  it('Loads the owner in template - hosted version tests', () => {
    expect(wrapperOwner.find('[data-test=owner]').text()).toEqual('Owner');
  });
  namespace.members.forEach((member) => {
    it(`Loads ${member.name} member in template`, () => {
      expect(wrapperOwner.find(`[data-test=${member.name}]`).text()).toEqual(member.name);
    });
  });
  it('Process data in methods - hosted version tests', () => {
    expect(wrapperOwner.vm.countDevicesHasNamespace()).toEqual(namespace.devices_count);

    let percent = 0;
    if (namespace.max_devices >= 0) {
      percent = (wrapperOwner.vm.countDevicesHasNamespace() / namespace.max_devices) * 100;
    }
    countDevicesHasNamespacePercent.maxDevices = namespace.max_devices;
    countDevicesHasNamespacePercent.percent = percent;

    expect(wrapperOwner.vm.countDevicesHasNamespacePercent())
      .toEqual(countDevicesHasNamespacePercent);
  });
  it('Check owner fields rendering in hosted version of the template - hosted version tests', () => {
    expect(wrapperOwner.find('[data-test=editOperation]').exists()).toEqual(true);
    expect(wrapperOwner.find('[data-test=userOperation]').exists()).toEqual(true);
    expect(wrapperOwner.find('[data-test=deleteOperation]').exists()).toEqual(true);
    expect(wrapperOwner.find('[data-test=securityOperation]').exists()).toEqual(true);
    expect(wrapperOwner.find('[data-test=notTheOwner]').exists()).toEqual(false);
    expect(wrapperOwner.findAll('[data-test=remove-member]').length).toEqual(namespace.members.length - 1);
    expect(wrapperOwner.find('[data-test=role]').exists()).toEqual(false);
    expect(wrapperOwner.find('[data-test=new-member]').exists()).toEqual(true);
  });
  it('Process data in methods - not the owner of this namespace', () => {
    expect(wrapperNotOwner.vm.hasTenant()).toEqual(false);
  });
  it('Check not the owner fields rendering in hosted version of the template - not the owner of this namespace', () => {
    const notTheOwnerMessage = 'You\'re not the owner of this namespace.';
    const namespaceOwnerMessage = `Contact ${namespace.members[0].name} user for more information.`;

    expect(wrapperNotOwner.find('[data-test=editOperation]').exists()).toEqual(false);
    expect(wrapperNotOwner.find('[data-test=userOperation]').exists()).toEqual(false);
    expect(wrapperNotOwner.find('[data-test=deleteOperation]').exists()).toEqual(false);
    expect(wrapperNotOwner.find('[data-test=securityOperation]').exists()).toEqual(false);
    expect(wrapperNotOwner.find('[data-test=notTheOwner]').exists()).toEqual(true);
    expect(wrapperNotOwner.find('[data-test=notTheOwner]').text()).toEqual(notTheOwnerMessage);
    expect(wrapperNotOwner.find('[data-test=namespaceOwnerMessage]').text()).toEqual(namespaceOwnerMessage);
  });
  it('Process data in the computed - open version tests', () => {
    expect(wrapperOwnerOpen.vm.isEnterprise).toEqual(false);
  });
  it('Process data in methods - open version tests', () => {
    expect(wrapperOwnerOpen.vm.countDevicesHasNamespace()).toEqual(openNamespace.devices_count);

    let percent = 0;
    if (openNamespace.max_devices >= 0) {
      percent = (wrapperOwnerOpen.vm.countDevicesHasNamespace() / openNamespace.max_devices) * 100;
    }
    countDevicesHasNamespacePercent.maxDevices = openNamespace.max_devices;
    countDevicesHasNamespacePercent.percent = percent;

    expect(wrapperOwnerOpen.vm.countDevicesHasNamespacePercent())
      .toEqual(countDevicesHasNamespacePercent);
  });
  it('Check owner fields rendering in open version of the template - open version tests', () => {
    expect(wrapperOwnerOpen.find('[data-test=editOperation]').exists()).toEqual(true);
    expect(wrapperOwnerOpen.find('[data-test=userOperation]').exists()).toEqual(true);
    expect(wrapperOwnerOpen.find('[data-test=deleteOperation]').exists()).toEqual(true);
    expect(wrapperOwnerOpen.find('[data-test=securityOperation]').exists()).toEqual(false);
    expect(wrapperOwnerOpen.find('[data-test=notTheOwner]').exists()).toEqual(false);
    expect(wrapperOwnerOpen.findAll('[data-test=remove-member]').exists()).toEqual(false);
    expect(wrapperOwnerOpen.find('[data-test=role]').exists()).toEqual(true);
    expect(wrapperOwnerOpen.findAll('[data-test=role]').wrappers.reduce((ac, v) => {
      ac.push(v.text());
      return ac;
    }, [])).toEqual(textRole);
    expect(wrapperOwnerOpen.find('[data-test=new-member]').exists()).toEqual(false);
  });
});
