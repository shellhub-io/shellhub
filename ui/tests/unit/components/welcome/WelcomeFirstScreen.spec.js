import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import WelcomeFirstScreen from '@/components/welcome/WelcomeFirstScreen';

describe('WelcomeFirstScreen', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const user = 'ShellHub';
  const name = 'ShellHub';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      user,
      name,
    },
    getters: {
      'auth/currentUser': (state) => state.user,
      'auth/currentName': (state) => state.name,
    },
    actions: {
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(WelcomeFirstScreen, {
      store,
      localVue,
      stubs: ['fragment'],
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

  it('Process data in the computed', () => {
    expect(wrapper.vm.name).toEqual(user);
  });
});
