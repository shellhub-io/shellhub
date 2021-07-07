import Vuex from 'vuex';
import { config, shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceAdd from '@/components/device/DeviceAdd';

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
  localVue.use(Vuex);

  let wrapper;

  const isOwner = true;
  const tenant = '00000000';

  const store = new Vuex.Store({
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
      'modals/showAddDevice': () => {
      },
      'snackbar/showSnackbarCopy': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(DeviceAdd, {
      store,
      localVue,
      stubs: ['fragment'],
      mocks: ['$copy', '$command'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.show).toBe(false);
    expect(wrapper.vm.tenant).toBe('00000000');
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
  it('Process data in methods', () => {
    const command = 'curl -sSf "http://localhost/install.sh?tenant_id=00000000" | sh';

    jest.spyOn(wrapper.vm, 'copyCommand');
    wrapper.vm.copyCommand();

    expect(wrapper.vm.command()).toBe(command);
    expect(wrapper.vm.copyCommand).toHaveBeenCalled();
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="command-field"]').exists()).toBe(true);
  });
});
