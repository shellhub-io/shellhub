import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import TokenEdit from '@/components/setting/token/TokenEdit';

describe('TokenEdit', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const token = {
    id: 'a582b47a42d',
    tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
    read_only: true,
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
    },
    getters: {
    },
    actions: {
      'tokens/put': () => {
      },
      'snackbar/showSnackbarSuccessAction': () => {
      },
      'snackbar/showSnackbarErrorAction': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(TokenEdit, {
      store,
      localVue,
      propsData: { token },
      stubs: ['fragment'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.token).toEqual(token);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
  });
});
