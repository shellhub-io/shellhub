import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import PrivateKeyFormDialogAdd from '@/components/private_key/PrivateKeyFormDialogAdd';
import '@/vee-validate';

describe('PrivateKeyFormDialogAdd', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const privateKey = {
    name: '',
    data: '',
  };

  const publicKey = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCxXq0NZ
  tbRBJlyyW5LOTMuqsZj3pL+Q5UCSQxnEjMpnz6yL6ALTS+fpVLzHIQwfZ3p5kMGk
  vAwXOwLuvkFpvQvbGsj7/kBov6zDeL7exdzPVvhVclsIU//aTm2ryT1898RFgEOm
  2YDSsNteG4dYBe9SbNJIbezAg7lCAdKxsbZD05phX8NewGOcFolPk8kSuYqJ6lWB
  /WWncLT8eXgP8Ew95rwug9Am3ApijuoD1j1RIb1LirF9xkNNg13DA9QYEFOO16XV
  EIxIS1frW7Krh+3LP2W6Q5ISFRzGF7hxlWs9RRzB/SG2WxrOpeQAoDOLrt/fu3g7
  sVL9pA32YbLgyAT`;

  const privateKeyRSA = `-----BEGIN RSA PRIVATE KEY-----
  MIIEowIBAAKCAQEAsV6tDWbW0QSZcsluSzkzLqrGY96S/kOVAkkMZxIzKZ8+si+g
  C00vn6VS8xyEMH2d6eZDBpLwMFzsC7r5Bab0L2xrI+/5AaL+sw3i+3sXcz1b4VXJ
  bCFP/2k5tq8k9fPfERYBDptmA0rDbXhuHWAXvUmzSSG3swIO5QgHSsbG2Q9OaYV/
  DXsBjnBaJT5PJErmKiepVgf1lp3C0/Hl4D/BMPea8LoPQJtwKYo7qA9Y9USG9S4q
  xfcZDTYNdwwPUGBBTjtel1RCMSEtX61uyq4ftyz9lukOSEhUcxhe4cZVrPUUcwf0
  htlsazqXkAKAzi67f37t4O7FS/aQN9mGy4MgEwIDAQABAoIBACm+XnwI+AW5T2P0
  hECv9ZvGFWrrtyygOzGOf5zCa8gf8mF9U+1U/SBViHAvBe1jowapapzheCXpuKQK
  HRF3eYCvw4dxsujvs4Hwgrss/RfiGP2vcvg/3tP9r9eO4OQBwT4EL7uAV0HvFo9t
  CH1hYDTsY4WSqek3UsoUWaL/pUzwKMijUgh2Dzj5o9AlNGWANu6txI1mIgHmwUvj
  2kV7E4R1mGynSprdsW68V36viB/V9d82XGxd3tYhKojiS1Dir68mR2U8ld3728Pd
  xU7o9x6NcWOtpTY1nS9MpufaYUTlp/chOXSd2RIY6JmtgbJcVTdE4rasfIAEnlAZ
  XALqKAECgYEA4kl6ZfcwKtxebVyczMCk6QWOJtsJ6CT17w2oehQGSuSLXeidjkFe
  bm04hUcN4Rm5iipUwDlA6JT8QoUgSG7Mjf8aDLv68FjXHxHjVvQaj0pg2I+1qADZ
  bN6m5xaazqAShF5MN4zQQTnNHTp6AIXOSQhIpqKS/Bjf3FYw48pxCyMCgYEAyKjf
  GnwiFJZN/q3s2mCmlEPblJ5mbXGCmIK/wjcoDST3+YrFi5VoWsHu0hRoZHtxIiaH
  sjSj8f8hWaZJ+yTL/V6zAO93JMovmoYyClmGt5pl56pFT2B7VGDC5FU9bylzWF3g
  HDdCTXOE72c5cOHnOddxVSBdD6GLC7Qe4CUVnlECgYBYNmSskywHyVhWMaA+gWrI
  HA5KP2EhSidFRYHD9UJut6FMvn2NExaI3bMG4agbdDfMEKxxMuCGym18UQFAu1Cq
  miPBixZL05Yo2oRRRV+FNG2EfqFGGO6pbjKKK1m16tjNGSWFEjOs+adoGX+t7Ht6
  JOyNaRr7g4bhEgiFBEoFGQKBgF3XI+dl8CZCmJ0nR6JlGuIxzen2Hh7Gu/WJCBbS
  5qcnB9UrAfGiYNg44/BZXOzJEgKPlFxR4+4Ti8w6SVTrQ37tn7crRkPtTk/svFA8
  yBTrXwb1iU5y55pxWhOgjYeEEg5ccKehbB9+i8fONX3GF/Xj/Ht8FClwOe+yP9JB
  ZZfRAoGBAJb08mFdb0Csbp+ed3LFznWINpXf2vlRKqIf+w8VOsEItbiB0r08AVdA
  Tik8VkRWm9ZHnMeMRRg2sEsI8gfaEXwSfLfMi10fn9YuWC2GSt5z+lA52H/S1zU2
  sGHPNn1H/cu7eM+nr9NxzJIT2CzKMHt5w4epp/UgkYFri4n2wDNS
  -----END RSA PRIVATE KEY-----`;

  const tests = [
    {
      description: 'Button create private Key',
      props: {
        show: false,
      },
      data: {
        privateKey,
        supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
      },
      template: {
        'createKey-btn': true,
        'privateKeyFormDialog-card': false,
      },
      templateText: {
        'createKey-btn': 'Add Private Key',
      },
    },
    {
      description: 'Dialog opened',
      props: {
        show: true,
      },
      data: {
        privateKey,
        supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
      },
      template: {
        'createKey-btn': true,
        'privateKeyFormDialog-card': true,
        'text-title': true,
        'name-field': true,
        'data-field': true,
        'cancel-btn': true,
        'create-btn': true,
      },
      templateText: {
        'createKey-btn': 'Add Private Key',
        'text-title': 'New Private Key',
        'name-field': '',
        'data-field': '',
        'cancel-btn': 'Cancel',
        'create-btn': 'Create',
      },
    },
  ];

  const storeVuex = () => new Vuex.Store({
    namespaced: true,
    state: { },
    getters: { },
    actions: {
      'privatekeys/set': () => {},
      'snackbar/showSnackbarSuccessNotRequest': () => {},
      'snackbar/showSnackbarErrorNotRequest': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(PrivateKeyFormDialogAdd, {
          store: storeVuex(),
          localVue,
          stubs: ['fragment'],
          propsData: {
            show: test.props.show,
          },
          vuetify,
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

      ///////
      // Data checking
      //////

      it('Compare data with default value', () => {
        Object.keys(test.data).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.data[item]);
        });
      });

      //////
      // HTML validation
      //////

      it('Renders the template with data', () => {
        Object.keys(test.template).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
        });
      });
      it('Renders template with expected text', () => {
        Object.keys(test.templateText).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).text()).toContain(test.templateText[item]);
        });
      });
      if (test.data.dialog) {
        it('Show validation messages', async () => {
          //////
          // In this case, the empty fields are validated.
          //////

          wrapper.setData({ privateKey: { name: '', data: '' } });
          await flushPromises();

          let validatorName = wrapper.vm.$refs.providerName;
          let validatorData = wrapper.vm.$refs.providerData;

          await validatorName.validate();
          await validatorData.validate();
          expect(validatorName.errors[0]).toBe('This field is required');
          expect(validatorData.errors[0]).toBe('This field is required');

          //////
          // In this case, any string is validated in the data.
          //////

          wrapper.setData({ privateKey: { data: 'xxxxxxxx' } });
          await flushPromises();

          validatorData = wrapper.vm.$refs.providerData;

          await validatorData.validate();
          expect(validatorData.errors[0]).toBe('Not valid key');

          //////
          // In this case, the public key is inserted where the private
          // key should be inserted.
          //////

          wrapper.setData({ privateKey: { data: publicKey } });
          await flushPromises();

          validatorData = wrapper.vm.$refs.providerData;

          await validatorData.validate();
          expect(validatorData.errors[0]).toBe('Not valid key');

          //////
          // In this case, the private key is inserted to validation.
          //////

          wrapper.setData({ privateKey: { name: 'ShellHub', data: privateKeyRSA } });
          await flushPromises();

          validatorName = wrapper.vm.$refs.providerName;
          validatorData = wrapper.vm.$refs.providerData;

          await validatorName.validate();
          await validatorData.validate();
          expect(validatorName.errors[0]).toBe(undefined);
          expect(validatorData.errors[0]).toBe(undefined);
        });
      }
    });
  });
});
