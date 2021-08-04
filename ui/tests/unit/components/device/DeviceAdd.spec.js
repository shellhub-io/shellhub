import Vuex from 'vuex';
import { config, mount, createLocalVue } from '@vue/test-utils';
import DeviceAdd from '@/components/device/DeviceAdd';
import Vuetify from 'vuetify';

// mocks global vars and clipboard function
config.mocks = {
  $copy: {
    command: 'Command',
    deviceSSHID: 'Device SSHID',
    tenantId: 'Tenant ID',
  },
  $clipboard: () => {},
};

describe('DeviceAdd', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const isOwner = true;
  const tenant = 'xxxxxxxx';

  const storeOwner = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner,
      addDevice: false,
      tenant,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
      'modals/addDevice': (state) => state.addDevice,
      'auth/tenant': (state) => state.tenant,
    },
    actions: {
      'modals/showAddDevice': () => {},
      'snackbar/showSnackbarCopy': () => {},
    },
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is button rendering.
  ///////

  describe('Button', () => {
    beforeEach(() => {
      wrapper = mount(DeviceAdd, {
        store: storeOwner,
        localVue,
        stubs: ['fragment'],
        mocks: ['$copy', '$command'],
        vuetify,
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

    it('Receive data in props', () => {
      expect(wrapper.vm.show).toBe(false);
      expect(wrapper.vm.tenant).toBe(tenant);
      expect(wrapper.vm.isOwner).toBe(isOwner);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.hostname).toEqual('localhost');
      expect(wrapper.vm.port).toEqual('');
      expect(wrapper.vm.dialog).toEqual(false);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.tenant).toEqual(tenant);
      expect(wrapper.vm.show).toEqual(false);
      expect(wrapper.vm.isOwner).toEqual(isOwner);
    });

    //////
    // HTML validation
    //////

    it('Process data in methods', () => {
      const command = 'curl -sSf "http://localhost/install.sh?tenant_id=xxxxxxxx" | sh';

      jest.spyOn(wrapper.vm, 'copyCommand');
      wrapper.vm.copyCommand();

      expect(wrapper.vm.command()).toBe(command);
      expect(wrapper.vm.copyCommand).toHaveBeenCalled();
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="add-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="command-field"]').exists()).toBe(false);
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is dialog rendering.
  ///////

  describe('Dialog', () => {
    beforeEach(() => {
      wrapper = mount(DeviceAdd, {
        store: storeOwner,
        localVue,
        stubs: ['fragment'],
        mocks: ['$copy', '$command'],
        vuetify,
      });

      wrapper.setData({ dialog: true });
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

    it('Receive data in props', () => {
      expect(wrapper.vm.show).toBe(false);
      expect(wrapper.vm.tenant).toBe(tenant);
      expect(wrapper.vm.isOwner).toBe(isOwner);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.hostname).toEqual('localhost');
      expect(wrapper.vm.port).toEqual('');
      expect(wrapper.vm.dialog).toEqual(true);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.tenant).toEqual(tenant);
      expect(wrapper.vm.show).toEqual(false);
      expect(wrapper.vm.isOwner).toEqual(isOwner);
    });

    //////
    // HTML validation
    //////

    it('Process data in methods', () => {
      const command = 'curl -sSf "http://localhost/install.sh?tenant_id=xxxxxxxx" | sh';

      jest.spyOn(wrapper.vm, 'copyCommand');
      wrapper.vm.copyCommand();

      expect(wrapper.vm.command()).toBe(command);
      expect(wrapper.vm.copyCommand).toHaveBeenCalled();
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="add-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="command-field"]').exists()).toBe(true);
    });
  });
});
