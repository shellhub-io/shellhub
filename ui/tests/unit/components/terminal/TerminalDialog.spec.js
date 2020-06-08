import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import TerminalDialog from '@/components/terminal/TerminalDialog.vue';

describe('TerminalDialog', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      terminal: '',
    },
    getters: {
      'modals/terminal': (state) => state.terminal,
    },
    actions: {
      'modals/toggleTerminal': () => {
      },
    }
  });

  beforeEach(() => {
    const uid = 'a582b47a42d';
    const username = 'user';
    const password = 'user';

    wrapper = shallowMount(TerminalDialog, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, username, password }
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper.isVueInstance()).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
