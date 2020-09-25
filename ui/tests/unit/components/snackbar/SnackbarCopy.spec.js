import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SnackbarCopy from '@/components/snackbar/SnackbarCopy';

describe('SnackbarCopy', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const mainContent = 'Command';
  const snackbarCopy = true;
  const message = `${mainContent} copied to clipboard.`;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      snackbarCopy,
    },
    getters: {
      'modals/snackbarCopy': (state) => state.snackbarCopy,
    },
    actions: {
      'modals/unsetShowStatusSnackbarCopy': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(SnackbarCopy, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { mainContent },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.mainContent).toEqual(mainContent);
  });
  it('Process data in the computed', async () => {
    expect(wrapper.vm.snackbar).toEqual(snackbarCopy);
    expect(wrapper.vm.message).toEqual(message);
  });
});
