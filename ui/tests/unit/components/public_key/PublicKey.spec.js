import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import publicKey from '@/components/public_key/PublicKey';
import Vuetify from 'vuetify';

describe('Session', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();

  let wrapper;

  const numberPublickeys = 0;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      numberPublickeys,
    },
    getters: {
      'publickeys/getNumberPublicKeys': (state) => state.numberPublickeys,
    },
    actions: {
      'publickeys/refresh': () => {
      },
      'boxs/setStatus': () => {
      },
      'publickeys/resetPagePerpage': () => {
      },
      'snackbar/showSnackbarErrorLoading': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = mount(publicKey, {
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
    expect(wrapper.vm.hasPublickey).toEqual(false);
  });
});
