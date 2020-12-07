import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import NamespaceDelete from '@/components/app_bar/namespace/NamespaceDelete';

describe('NamespaceDelete', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const tenant = 'xxxxxx';

  const store = new Vuex.Store({
    namespaced: true,
    actions: {
      'namespaces/remove': () => {
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
