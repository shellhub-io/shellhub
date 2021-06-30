import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Session from '@/components/session/Session';
import Vuetify from 'vuetify';

describe('Session', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const numberSessions = 0;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      numberSessions,
    },
    getters: {
      'sessions/getNumberSessions': (state) => state.numberSessions,
    },
    actions: {
      'sessions/refresh': () => {
      },
      'boxs/setStatus': () => {
      },
      'sessions/resetPagePerpage': () => {
      },
      'snackbar/showSnackbarErrorLoading': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = mount(Session, {
      store,
      localVue,
      stubs: ['fragment'],
      vuetify,
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.hasSession).toEqual(false);
  });
});
