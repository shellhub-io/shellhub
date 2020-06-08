import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DetailsSession from '@/views/DetailsSession.vue';

describe('Terminal', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  beforeEach(() => {
    wrapper = shallowMount(DetailsSession, {
      localVue,
      stubs: ['fragment'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper.isVueInstance()).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
