import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import TerminalDialog from '@/components/terminal/TerminalDialog';

describe('TerminalDialog', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const uid = 'a582b47a42d';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      terminal: 'a582b47a42d',
    },
    getters: {
      'modals/terminal': (state) => state.terminal,
    },
    actions: {
      'modals/toggleTerminal': () => {
      },
    },
  });

  beforeEach(() => {
    const username = 'user';
    const password = 'user';

    wrapper = shallowMount(TerminalDialog, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, username, password },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Compare data with default value', async () => {
    expect(wrapper.vm.username).toEqual('');
    expect(wrapper.vm.passwd).toEqual('');
    expect(wrapper.vm.showLoginForm).toEqual(true);
    expect(wrapper.vm.valid).toEqual(true);
    await wrapper.setData({ rules: { required: 'Required' } });
    expect(wrapper.vm.rules).toEqual({ required: 'Required' });
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.uid).toEqual(uid);
  });
  it('Receive data in computed', () => {
    expect(wrapper.vm.show).toEqual(true);
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="username"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="passwd"]').exists()).toBe(true);
  });
});
