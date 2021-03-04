import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import BoxMessage from '@/components/box/BoxMessage';
import Vuetify from 'vuetify';

describe('BoxMessage', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();

  let wrapper;

  const typeMessage = {
    device: 'device',
    session: 'session',
    firewall: 'firewall',
    publicKey: 'publicKey',
  };

  const items = {
    device:
    {
      icon: 'devices',
      title: 'Device',
      text: [
        'In order to register a device on ShellHub, you need to install ShellHub agent onto it.',
      ],
      textWithLink: [
        `The easiest way to install ShellHub agent is with our automatic one-line installation
          script, which works with all Linux distributions that have Docker installed and
          properly set up See More.`,
      ],
    },
    session:
    {
      icon: 'history',
      title: 'Session',
      text: [
        'An SSH session is created when a connection is made to any registered device.',
      ],
      textWithLink: [
        `Please follow our guide on how to connect to your devices
            See More.`,
      ],
    },
    firewall:
    {
      icon: 'security',
      title: 'Firewall Rule',
      text: [
        `ShellHub provides flexible firewall for filtering SSH connections.
              It gives a fine-grained control over which SSH connections reach the devices.`,
        `Using Firewall Rules you can deny or allow SSH connections from specific
              IP address to a specific or a group of devices using a given username.`,
      ],
      textWithLink: [],
    },
    publicKey:
    {
      icon: 'vpn_key',
      title: 'Public Keys',
      text: [
        'You can connect to your devices using password-based logins, but we strongly recommend using SSH key pairs instead.',
        'SSH keys are more secure than passwords and can help you log in without having to remember long passwords.',
      ],
      textWithLink: [],
    },
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
    },
    getters: {
    },
    actions: {
      'boxs/setStatus': () => {
      },
      'firewallrules/refresh': () => {
      },
      'publickeys/refresh': () => {
      },
      'snackbar/showSnackbarErrorLoading': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(BoxMessage, {
      store,
      localVue,
      stubs: ['fragment'],
      vuetify,
      propsData: { typeMessage: typeMessage.session },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Process data in methods Device', () => {
    const deviceWrapper = shallowMount(BoxMessage, {
      store,
      localVue,
      stubs: ['fragment'],
      vuetify,
      propsData: { typeMessage: typeMessage.device },
    });

    const title = `Looks like you don't have any ${items.device.title}`;

    expect(deviceWrapper.find('[data-test="boxMessage-icon"]').text()).toEqual(items.device.icon);
    expect(deviceWrapper.find('[data-test="boxMessage-title"]').text()).toEqual(title);
    Object.keys(items.device.text).forEach((index) => {
      expect(deviceWrapper.find(`[data-test="${index}-boxMessage-text"]`).text()).toEqual(items.device.text[index]);
    });

    const lenDeviceText = (items.device.text).length;
    Object.keys(items.firewall.textWithLink).forEach((index) => {
      expect(deviceWrapper.find(`[data-test="${lenDeviceText + parseInt(index, 10)}-boxMessage-text"]`).text()).toEqual(items.device.textWithLink[index]);
    });
  });
  it('Process data in methods Session', () => {
    const sessionWrapper = shallowMount(BoxMessage, {
      store,
      localVue,
      stubs: ['fragment'],
      vuetify,
      propsData: { typeMessage: typeMessage.session },
    });

    const title = `Looks like you don't have any ${items.session.title}`;

    expect(sessionWrapper.find('[data-test="boxMessage-icon"]').text()).toEqual(items.session.icon);
    expect(sessionWrapper.find('[data-test="boxMessage-title"]').text()).toEqual(title);
    Object.keys(items.session.text).forEach((index) => {
      expect(sessionWrapper.find(`[data-test="${index}-boxMessage-text"]`).text()).toEqual(items.session.text[index]);
    });

    const lenSessionText = (items.session.text).length;
    Object.keys(items.session.textWithLink).forEach((index) => {
      expect(sessionWrapper.find(`[data-test="${lenSessionText + parseInt(index, 10)}-boxMessage-text"]`).text()).toEqual(items.session.textWithLink[index]);
    });
  });
  it('Process data in methods FirewallRule', () => {
    const firewallWrapper = shallowMount(BoxMessage, {
      store,
      localVue,
      stubs: ['fragment'],
      vuetify,
      propsData: { typeMessage: typeMessage.firewall },
    });

    const title = `Looks like you don't have any ${items.firewall.title}`;

    expect(firewallWrapper.find('[data-test="boxMessage-icon"]').text()).toEqual(items.firewall.icon);
    expect(firewallWrapper.find('[data-test="boxMessage-title"]').text()).toEqual(title);
    Object.keys(items.firewall.text).forEach((index) => {
      expect(firewallWrapper.find(`[data-test="${index}-boxMessage-text"]`).text()).toEqual(items.firewall.text[index]);
    });

    const lenFirewallText = (items.firewall.text).length;
    Object.keys(items.firewall.textWithLink).forEach((index) => {
      expect(firewallWrapper.find(`[data-test="${lenFirewallText + parseInt(index, 10)}-boxMessage-text"]`).text()).toEqual(items.firewall.textWithLink[index]);
    });
  });
  it('Process data in methods Public Key', () => {
    const publicKeyWrapper = shallowMount(BoxMessage, {
      store,
      localVue,
      stubs: ['fragment'],
      vuetify,
      propsData: { typeMessage: typeMessage.publicKey },
    });

    const title = `Looks like you don't have any ${items.publicKey.title}`;

    expect(publicKeyWrapper.find('[data-test="boxMessage-icon"]').text()).toEqual(items.publicKey.icon);
    expect(publicKeyWrapper.find('[data-test="boxMessage-title"]').text()).toEqual(title);
    Object.keys(items.publicKey.text).forEach((index) => {
      expect(publicKeyWrapper.find(`[data-test="${index}-boxMessage-text"]`).text()).toEqual(items.publicKey.text[index]);
    });

    const lenPublicKeyText = (items.publicKey.text).length;
    Object.keys(items.firewall.textWithLink).forEach((index) => {
      expect(publicKeyWrapper.find(`[data-test="${lenPublicKeyText + parseInt(index, 10)}-boxMessage-text"]`).text()).toEqual(items.publicKey.textWithLink[index]);
    });
  });
});
