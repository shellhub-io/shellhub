import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SessionPlay from '@/components/session/SessionPlay';

describe('SessionPlay', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  const uid = 'a582b47a42d';
  const recorded = true;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      logSession: [],
    },
    getters: {
      'sessions/getLogSession': (state) => state.logSession,
    },
    actions: {
      'sessions/getLogSession': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(SessionPlay, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, recorded },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
