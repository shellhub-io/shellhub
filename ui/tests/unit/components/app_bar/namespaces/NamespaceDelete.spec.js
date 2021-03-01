import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import NamespaceDelete from '@/components/app_bar/namespace/NamespaceDelete';

describe('NamespaceDelete', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const tenant = 'xxxxxx';

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
      'namespaces/remove': () => {
      },
      'auth/logout': () => {
      },
      'snackbar/showSnackbarErrorLoading': () => {
      },
      'snackbar/showSnackbarSuccessAction': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(NamespaceDelete, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { nsTenant: tenant },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receives data in props', () => {
    expect(wrapper.vm.nsTenant).toEqual(tenant);
  });
});
