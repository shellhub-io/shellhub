import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import WelcomeSecondScreen from '@/components/welcome/WelcomeSecondScreen';

describe('WelcomeSecondScreen', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const command = '';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
    },
    getters: {
    },
    actions: {
      'modals/showSnackbarCopy': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(WelcomeSecondScreen, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { command },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
