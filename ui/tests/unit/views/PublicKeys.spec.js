import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Publickeys from '@/views/PublicKeys';

describe('Publickeys', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  beforeEach(() => {
    wrapper = shallowMount(Publickeys, {
      localVue,
      stubs: ['fragment'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
