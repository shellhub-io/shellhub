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
    members: [{ name: 'user6' }, { name: 'user7' }, { name: 'user8' }],
    tenant_id: 'e359bf484715',
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
    jest.spyOn(Storage.prototype, 'getItem').mockReturnValue('e359bf484715');

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
  it('Process data in the computed', () => {
    expect(wrapper.vm.namespace).toEqual(namespace);
    expect(wrapper.vm.isHosted).toEqual(false);
  });
  it('Loads name when component is created', () => {
    wrapper.vm.$nextTick(() => {
      expect(wrapper.vm.name).toBe(namespace.name);
    });
  });
  namespace.members.forEach((member) => {
    it(`Loads ${member} member in template`, () => {
      expect(wrapper.find(`[data-test=${member.name}]`).text()).toEqual(member.name);
    });
  });
});
