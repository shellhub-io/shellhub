import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Login from '@/views/Login.vue';

const localVue = createLocalVue();
localVue.use(Vuex);

const store = new Vuex.Store({
  state: {
    auth: {
      status: true,
      token: 'akjhsdkjahsd',
      user: 'leonardojoao',
      tenant: 'kldsjflksjdfkl'
    }
  },
  getters: {
    getCurrentLocation: (state) => state.currentLocation
  }
});

describe('Header', () => {
  const wrapper = shallowMount(Login, {
    store,
    localVue
  });
  it('is a Vue instance', () => {
    expect(wrapper.isVueInstance()).toBeTruthy();
  });

});
