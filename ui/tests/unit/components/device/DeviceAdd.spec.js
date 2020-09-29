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

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      addDevice: false,
      tenant: '00000000',
    },
    getters: {
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
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="command-field"]').exists()).toBe(true);
  });
  it('Proccess computed data in computed', () => {
    expect(wrapper.vm.show).toBe(false);
    expect(wrapper.vm.tenant).toBe('00000000');
  });
  it('Process data in methods', () => {
    const command = 'curl "http://localhost/install.sh?tenant_id=00000000" | sh';
    expect(wrapper.vm.command()).toBe(command);

    jest.spyOn(wrapper.vm, 'copyCommand');
    wrapper.vm.copyCommand();
    expect(wrapper.vm.copyCommand).toHaveBeenCalled();
  });
});
