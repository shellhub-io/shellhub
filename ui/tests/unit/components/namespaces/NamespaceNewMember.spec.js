import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import NamespaceNewMember from '@/components/namespace/NamespaceNewMember';

describe('NamespaceNewMember', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  const tenant = 'xxxxx';

  let wrapper;

  const store = new Vuex.Store({
    namespaced: true,
    actions: {
      'namespaces/adduser': () => {
      },
      'snackbar/showSnackbarSuccessAction': () => {
      },
      'snackbar/showSnackbarErrorAction': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(NamespaceNewMember, {
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
  it('Receive data in props', () => {
    expect(wrapper.vm.nsTenant).toEqual(tenant);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.tenant).toEqual(tenant);
  });
});
