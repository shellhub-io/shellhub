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

  let wrapper;

  const idOwner = '6';
  const idNotOwner = '10';
  const owner = true;
  const hasTenant = true;
  const isEnterpriseOwner = true;
  const isEnterprise = true;
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
      'namespaces/put': () => {},
      'namespaces/get': () => {},
      'namespaces/removeUser': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
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
      'namespaces/put': () => {},
      'namespaces/get': () => {},
      'namespaces/removeUser': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
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
      'namespaces/put': () => {},
      'namespaces/get': () => {},
      'namespaces/removeUser': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  ///////
  // In this case, hosted version tests
  ///////

  describe('Hosted version', () => {
    beforeEach(() => {
      jest.spyOn(Storage.prototype, 'getItem').mockReturnValue('e359bf484715');

      wrapper = shallowMount(SettingNamespace, {
        store: storeOwner,
        localVue,
        stubs: ['fragment'],
        mocks: ['$env'],
      });
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue instance', () => {
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Compare data with default value', () => {
      expect(wrapper.vm.name).toEqual(namespace.name);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.isOwner).toEqual(owner);
      expect(wrapper.vm.owner).toEqual(namespace.owner);
      expect(wrapper.vm.namespace).toEqual(namespace);
      expect(wrapper.vm.tenant).toEqual(namespace.tenant_id);
      expect(wrapper.vm.isEnterpriseOwner).toEqual(isEnterpriseOwner);
      expect(wrapper.vm.isEnterprise).toEqual(isEnterprise);
    });
    it('Process data in methods', () => {
      let percent = 0;
      if (namespace.max_devices >= 0) {
        percent = (wrapper.vm.countDevicesHasNamespace() / namespace.max_devices) * 100;
      }
      countDevicesHasNamespacePercent.maxDevices = namespace.max_devices;
      countDevicesHasNamespacePercent.percent = percent;

      expect(wrapper.vm.hasTenant()).toEqual(hasTenant);
      expect(wrapper.vm.countDevicesHasNamespace()).toEqual(namespace.devices_count);
      expect(wrapper.vm.countDevicesHasNamespacePercent())
        .toEqual(countDevicesHasNamespacePercent);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="tenant-span"]').text()).toEqual(namespace.tenant_id);

      //////
      // Check rendering of member names in the list.
      //////

      namespace.members.forEach((member) => {
        expect(wrapper.find(`[data-test="${member.name}-list"]`).text()).toEqual(member.name);
      });

      //////
      // Check owner fields rendering.
      //////

      expect(wrapper.find('[data-test="owner-p"]').text()).toEqual('Owner');

      expect(wrapper.find('[data-test="editOperation-div"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="userOperation-div"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="deleteOperation-div"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="securityOperation-div"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="notTheOwner-span"]').exists()).toEqual(false);
      expect(wrapper.findAll('[data-test="removeMember-btn"]').length).toEqual(namespace.members.length - 1);
      expect(wrapper.find('[data-test="role-div"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="newMember-div"]').exists()).toEqual(true);
    });
  });

  ///////
  // In this case, not the owner of this namespace
  ///////

  describe('Not the owner of this namespace', () => {
    beforeEach(() => {
      jest.spyOn(Storage.prototype, 'getItem').mockReturnValue('e359bf484715');

      wrapper = shallowMount(SettingNamespace, {
        localVue,
        store: storeNotOwner,
        stubs: ['fragment'],
        mocks: ['$env'],
      });
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue instance', () => {
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Compare data with default value', () => {
      expect(wrapper.vm.name).toEqual('');
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.isOwner).toEqual(!owner);
      expect(wrapper.vm.owner).toEqual(namespace.owner);
      expect(wrapper.vm.namespace).toEqual(namespace);
      expect(wrapper.vm.tenant).toEqual('');
      expect(wrapper.vm.isEnterpriseOwner).toEqual(!isEnterpriseOwner);
      expect(wrapper.vm.isEnterprise).toEqual(isEnterprise);
    });
    it('Process data in methods', () => {
      let percent = 0;
      if (namespace.max_devices >= 0) {
        percent = (wrapper.vm.countDevicesHasNamespace() / namespace.max_devices) * 100;
      }
      countDevicesHasNamespacePercent.maxDevices = namespace.max_devices;
      countDevicesHasNamespacePercent.percent = percent;

      expect(wrapper.vm.hasTenant()).toEqual(!hasTenant);
      expect(wrapper.vm.countDevicesHasNamespace()).toEqual(namespace.devices_count);
      expect(wrapper.vm.countDevicesHasNamespacePercent())
        .toEqual(countDevicesHasNamespacePercent);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="tenant-span"]').text()).toEqual('');

      //////
      // Check owner fields rendering.
      //////
      const namespaceOwnerMessage = `Contact ${namespace.members[0].name} user for more information.`;

      expect(wrapper.find('[data-test="owner-p"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="editOperation-div"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="userOperation-div"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="deleteOperation-div"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="securityOperation-div"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="notTheOwner-span"]').exists()).toEqual(true);
      expect(wrapper.findAll('[data-test="removeMember-btn"]').length).toEqual(0);
      expect(wrapper.find('[data-test="role-div"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="newMember-div"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test=namespaceOwnerMessage-p]').text()).toEqual(namespaceOwnerMessage);
    });
  });

  ///////
  // In this case, open version tests
  ///////

  describe('Open version', () => {
    beforeEach(() => {
      jest.spyOn(Storage.prototype, 'getItem').mockReturnValue('e359bf484715');

      wrapper = shallowMount(SettingNamespace, {
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

    ///////
    // Component Rendering
    //////

    it('Is a Vue instance', () => {
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Compare data with default value', () => {
      expect(wrapper.vm.name).toEqual(namespace.name);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.isOwner).toEqual(owner);
      expect(wrapper.vm.owner).toEqual(namespace.owner);
      expect(wrapper.vm.namespace).toEqual({ ...namespace, max_devices: -1 });
      expect(wrapper.vm.tenant).toEqual(namespace.tenant_id);
      expect(wrapper.vm.isEnterpriseOwner).toEqual(!isEnterpriseOwner);
      expect(wrapper.vm.isEnterprise).toEqual(!isEnterprise);
    });
    it('Process data in methods', () => {
      let percent = 0;
      if (namespace.max_devices >= 0) {
        percent = (wrapper.vm.countDevicesHasNamespace() / namespace.max_devices) * 100;
      }
      countDevicesHasNamespacePercent.maxDevices = namespace.max_devices;
      countDevicesHasNamespacePercent.percent = percent;

      expect(wrapper.vm.hasTenant()).toEqual(hasTenant);
      expect(wrapper.vm.countDevicesHasNamespace()).toEqual(namespace.devices_count);
      expect(wrapper.vm.countDevicesHasNamespacePercent())
        .toEqual({ ...countDevicesHasNamespacePercent, maxDevices: -1, percent: 0 });
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="tenant-span"]').text()).toEqual(namespace.tenant_id);

      //////
      // Check rendering of member names in the list.
      //////

      namespace.members.forEach((member) => {
        expect(wrapper.find(`[data-test="${member.name}-list"]`).text()).toEqual(member.name);
      });

      //////
      // Check owner fields rendering.
      //////

      expect(wrapper.find('[data-test="owner-p"]').exists()).toEqual(false);

      expect(wrapper.find('[data-test="editOperation-div"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="userOperation-div"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="deleteOperation-div"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="securityOperation-div"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="notTheOwner-span"]').exists()).toEqual(false);
      expect(wrapper.findAll('[data-test="removeMember-btn"]').exists()).toEqual(false);
      expect(wrapper.findAll('[data-test=role-div]').wrappers.reduce((ac, v) => {
        ac.push(v.text());
        return ac;
      }, [])).toEqual(textRole);
      expect(wrapper.find('[data-test="newMember-div"]').exists()).toEqual(false);
    });
  });
});
