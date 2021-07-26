import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Sessions from '@/views/Sessions';

describe('Sessions', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  beforeEach(() => {
    wrapper = shallowMount(Sessions, {
      localVue,
      stubs: ['fragment'],
    });
  });

  ///////
  // Component Rendering
  //////

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  //////
  // HTML validation
  //////

  it('Renders the template with components', () => {
    expect(wrapper.find('[data-test="session-component"]').exists()).toBe(true);
  });
});
