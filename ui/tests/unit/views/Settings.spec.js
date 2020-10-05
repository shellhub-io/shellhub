import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import VueRouter from 'vue-router';
import Settings from '@/views/Settings';

describe('Settings', () => {
  const localVue = createLocalVue();
  localVue.use(VueRouter);
  localVue.use(Vuex);

  let wrapper;

  beforeEach(() => {
    wrapper = shallowMount(Settings, {
      localVue,
      stubs: ['fragment'],
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
