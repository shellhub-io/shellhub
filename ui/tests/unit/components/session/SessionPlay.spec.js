import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SessionPlay from '@/components/session/SessionPlay';

describe('SessionPlay', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const uid = '8c354a00f50';
  const recorded = true;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      session: [],
    },
    getters: {
      'sessions/get': (state) => state.session,
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
      mocks: {
        $env: (isHosted) => isHosted,
      },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
