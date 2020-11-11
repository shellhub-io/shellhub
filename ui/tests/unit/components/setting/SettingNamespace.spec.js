import Vuex from 'vuex';
import { shallowMount, createLocalVue, config } from '@vue/test-utils';
import SettingNamespace from '@/components/setting/SettingNamespace';

config.mocks = {
  $env: {
    isHosted: true,
  },
};

describe('SettingNamespace', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const namespace = {
    name: 'namespace3',
    owner: 'user1',
    member_names: ['user6', 'user7', 'user8'],
    tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484715',
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
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
    wrapper = shallowMount(SettingNamespace, {
      store,
      localVue,
      stubs: ['fragment'],
      mocks: ['$env'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.name).toEqual('');
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.namespace).toEqual(namespace);
    expect(wrapper.vm.show).toEqual(true);
  });
  it('Loads name when component is created', () => {
    wrapper.vm.$nextTick(() => {
      expect(wrapper.vm.name).toBe(namespace.name);
    });
  });
  namespace.member_names.forEach((member) => {
    it(`Loads ${member} member in template`, () => {
      expect(wrapper.find(`[data-test=${member}]`).text()).toEqual(member);
    });
  });
});
