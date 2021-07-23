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
      'snackbar/snackbarCopy': (state) => state.snackbarCopy,
    },
    actions: {
      'snackbar/unsetShowStatusSnackbarCopy': () => {},
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
    expect(wrapper.vm.mainContent).toEqual(mainContent);
  });
  it('Process data in the computed', async () => {
    expect(wrapper.vm.snackbar).toEqual(snackbarCopy);
    expect(wrapper.vm.message).toEqual(message);
  });

  //////
  // HTML validation
  //////

  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="message-snackbar"]').text()).toEqual(message);
  });
});
