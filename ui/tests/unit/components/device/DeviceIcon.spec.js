import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceIcon from '@/components/device/DeviceIcon.vue';

describe('DeviceIcon', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  let iconName = 'alpine';

  const iconsMap = {
    alpine: 'fl-alpine',
    arch: 'fl-archlinux',
    centos: 'fl-centos',
    coreos: 'fl-coreos',
    debian: 'fl-debian',
    devuan: 'fl-devuan',
    elementary: 'fl-elementary',
    fedora: 'fl-fedora',
    freebsd: 'fl-freebsd',
    gentoo: 'fl-gentoo',
    linuxmint: 'fl-linuxmint',
    mageia: 'fl-mageia',
    manjaro: 'fl-manjaro',
    mandriva: 'fl-mandriva',
    nixos: 'fl-nixos',
    opensuse: 'fl-opensuse',
    rhel: 'fl-redhat',
    sabayon: 'fl-sabayon',
    slackware: 'fl-slackware',
    ubuntu: 'fl-ubuntu',
    raspbian: 'fl-raspberry-pi',
    'ubuntu-core': 'fl-ubuntu',
    void: 'fl-void',
    default: 'fl-tux'
  };

  beforeEach(() => {

    wrapper = shallowMount(DeviceIcon, {
      localVue,
      stubs: ['fragment'],
      propsData: { iconName }
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper.isVueInstance()).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  for (iconName in iconsMap) {
    it(`Has the ${iconName} icon`, () => {
      wrapper = shallowMount(DeviceIcon, {
        localVue,
        stubs: ['fragment'],
        propsData: { iconName }
      });
      expect(wrapper.find('[data-cy="type-icon"]').text()).toBe(iconsMap[iconName]);
    });
  }
});
