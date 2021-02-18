import { mount, createLocalVue } from '@vue/test-utils';
import BoxMessage from '@/components/box/BoxMessage';
import Vuetify from 'vuetify';

describe('BoxMessage', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();

  let wrapper;

  const typeMessage = 'firewall';

  const items = {
    firewall:
    {
      icon: 'security',
      title: 'Firewall Rule',
      text: [`ShellHub provides flexible firewall for filtering SSH connections.
            It gives a fine-grained control over which SSH connections reach the devices.`,
      `Using Firewall Rules you can deny or allow SSH connections from specific
            IP address to a specific or a group of devices using a given username.`],
    },
  };

  beforeEach(() => {
    wrapper = mount(BoxMessage, {
      localVue,
      stubs: ['fragment'],
      vuetify,
      propsData: { typeMessage },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Process data in methods', () => {
    const title = `Looks like you don't have any ${items.firewall.title}`;

    expect(wrapper.find('[data-test="boxMessage-icon"]').text()).toEqual(items.firewall.icon);
    expect(wrapper.find('[data-test="boxMessage-title"]').text()).toEqual(title);
    Object.keys(items.firewall.text).forEach((index) => {
      expect(wrapper.find(`[data-test="${index}-boxMessage-text"]`).text()).toEqual(items.firewall.text[index]);
    });
  });
});
