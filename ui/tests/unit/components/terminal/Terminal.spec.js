import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Terminal from '@/components/terminal/Terminal.vue';

describe('Terminal', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  beforeEach(() => {
    const uid = 'a582b47a42d';
    const username = 'user';
    const password = 'user';

    wrapper = shallowMount(Terminal, {
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
