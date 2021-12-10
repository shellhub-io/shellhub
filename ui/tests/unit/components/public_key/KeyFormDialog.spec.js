import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import KeyFormDialog from '@/components/public_key/KeyFormDialog';
import { actions, authorizer } from '../../../../src/authorizer';
import '@/vee-validate';

describe('KeyFormDialog', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const role = ['owner', 'administrator', 'operator', 'observer'];

  const hasAuthorization = {
    owner: true,
    administrator: true,
    operator: false,
    observer: false,
  };

  const privateKeyLocal = {
    data: '',
    name: '',
  };

  const publicKeyLocal = { ...privateKeyLocal, hostname: '' };

  const keyObject = {
    data: 'AbGVvbmF',
    fingerprint: '00:00:00',
    created_at: '2020-11-23T20:59:13.323Z',
    tenant_id: 'xxxxxxxx',
    name: 'shellhub',
  };

  // The keys were generated from the website:
  // https://www.devglan.com/online-tools/rsa-encryption-decryption

  const publicKey = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCxXq0NZ
  tbRBJlyyW5LOTMuqsZj3pL+Q5UCSQxnEjMpnz6yL6ALTS+fpVLzHIQwfZ3p5kMGk
  vAwXOwLuvkFpvQvbGsj7/kBov6zDeL7exdzPVvhVclsIU//aTm2ryT1898RFgEOm
  2YDSsNteG4dYBe9SbNJIbezAg7lCAdKxsbZD05phX8NewGOcFolPk8kSuYqJ6lWB
  /WWncLT8eXgP8Ew95rwug9Am3ApijuoD1j1RIb1LirF9xkNNg13DA9QYEFOO16XV
  EIxIS1frW7Krh+3LP2W6Q5ISFRzGF7hxlWs9RRzB/SG2WxrOpeQAoDOLrt/fu3g7
  sVL9pA32YbLgyAT`;

  const privateKey = `-----BEGIN RSA PRIVATE KEY-----
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
      description: 'Button create publicKey',
      variables: {
        createKey: true,
        dialog: false,
      },
      props: {
        keyObject: {},
        createKey: true,
        action: 'public',
      },
      data: {
        dialog: false,
        keyLocal: publicKeyLocal,
        supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
      },
      template: {
        'createKey-btn': true,
        'keyFormDialog-card': false,
      },
    },
    {
      description: 'Icon edit publicKey',
      variables: {
        createKey: false,
        dialog: false,
      },
      props: {
        keyObject,
        createKey: false,
        action: 'public',
      },
      data: {
        dialog: false,
        supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
      },
      template: {
        'createKey-btn': false,
        'keyFormDialog-card': false,
      },
    },
    {
      description: 'Button create privateKey',
      variables: {
        createKey: true,
        dialog: false,
      },
      props: {
        keyObject: {},
        createKey: true,
        action: 'private',
      },
      data: {
        dialog: false,
        keyLocal: privateKeyLocal,
        supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
      },
      template: {
        'createKey-btn': true,
        'keyFormDialog-card': false,
      },
    },
    {
      description: 'Icon edit privateKey',
      variables: {
        createKey: false,
        dialog: false,
      },
      props: {
        keyObject: { name: 'xxxxxxx', data: 'xxxxxxx' },
        createKey: false,
        action: 'private',
      },
      data: {
        dialog: false,
        keyLocal: { name: 'xxxxxxx', data: 'xxxxxxx' },
        supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
      },
      template: {
        'createKey-btn': false,
        'keyFormDialog-card': false,
      },
    },
    {
      description: 'Dialog edit publicKey',
      variables: {
        createKey: false,
        dialog: true,
      },
      props: {
        keyObject,
        createKey: false,
        action: 'public',
      },
      data: {
        dialog: true,
        supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
      },
      template: {
        'createKey-btn': false,
        'keyFormDialog-card': true,
      },
    },
    {
      description: 'Dialog edit privateKey',
      variables: {
        createKey: false,
        dialog: true,
      },
      props: {
        keyObject: { name: 'xxxxxxx', data: 'xxxxxxx' },
        createKey: false,
        action: 'private',
      },
      data: {
        dialog: true,
        keyLocal: { name: 'xxxxxxx', data: 'xxxxxxx' },
        supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
      },
      template: {
        'createKey-btn': false,
        'keyFormDialog-card': true,
      },
    },
  ];

  const storeVuex = (currentrole) => new Vuex.Store({
    namespaced: true,
    state: {
      currentrole,
    },
    getters: {
      'auth/role': (state) => state.currentrole,
    },
    actions: {
      'publickeys/post': () => {},
      'publickeys/put': () => {},
      'privatekeys/set': () => {},
      'privatekeys/edit': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'snackbar/showSnackbarSuccessNotRequest': () => {},
      'snackbar/showSnackbarErrorNotRequest': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = mount(KeyFormDialog, {
            store: storeVuex(currentrole),
            localVue,
            stubs: ['fragment'],
            propsData: {
              keyObject: test.props.keyObject,
              createKey: test.props.createKey,
              action: test.props.action,
            },
            vuetify,
            mocks: {
              $authorizer: authorizer,
              $actions: actions,
            },
          });

          wrapper.setData({ dialog: test.variables.dialog });
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

        it('Receive data in props', () => {
          Object.keys(test.props).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.props[item]);
          });
        });
        it('Compare data with default value', () => {
          Object.keys(test.data).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.data[item]);
          });
        });
        it('Process data in the computed', () => {
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentrole]);
        });

        //////
        // HTML validation
        //////

        it('Renders the template with data', () => {
          Object.keys(test.template).forEach((item) => {
            expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
          });
        });

        if (!test.data.dialog) {
          if (hasAuthorization[currentrole] && !test.variables.createKey) {
            it('Show message tooltip to user owner', async (done) => {
              const icons = wrapper.findAll('.v-icon');
              const helpIcon = icons.at(0);
              helpIcon.trigger('mouseenter');
              await wrapper.vm.$nextTick();

              expect(icons.length).toBe(1);
              expect(helpIcon.text()).toEqual('edit');
              requestAnimationFrame(() => {
                expect(wrapper.find('[data-test="text-tooltip"]').text()).toEqual('Edit');
                done();
              });
            });
          }
        } else if (test.props.action === 'public') {
          it('Show validation messages', async () => {
            //////
            // In this case, the empty fields are validated.
            //////

            wrapper.setData({ keyLocal: { name: '', data: '' } });
            await flushPromises();

            const validatorName = wrapper.vm.$refs.providerName;
            let validatorData = wrapper.vm.$refs.providerData;

            await validatorName.validate();
            await validatorData.validate();
            expect(validatorName.errors[0]).toBe('This field is required');
            expect(validatorData.errors[0]).toBe('This field is required');

            //////
            // In this case, any string is validated in the data.
            //////

            wrapper.setData({ keyLocal: { data: 'xxxxxxxx' } });
            await flushPromises();

            validatorData = wrapper.vm.$refs.providerData;

            await validatorData.validate();
            expect(validatorData.errors[0]).toBe('Not valid key');
          });
        } else if (test.props.action === 'private') {
          it('Show validation messages', async () => {
            //////
            // In this case, the empty fields are validated.
            //////

            wrapper.setData({ keyLocal: { name: '', data: '' } });
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

            wrapper.setData({ keyLocal: { data: 'xxxxxxxx' } });
            await flushPromises();

            validatorData = wrapper.vm.$refs.providerData;

            await validatorData.validate();
            expect(validatorData.errors[0]).toBe('Not valid key');

            //////
            // In this case, the public key is inserted where the private
            // key should be inserted.
            //////

            wrapper.setData({ keyLocal: { data: publicKey } });
            await flushPromises();

            validatorData = wrapper.vm.$refs.providerData;

            await validatorData.validate();
            expect(validatorData.errors[0]).toBe('Not valid key');

            //////
            // In this case, the private key is inserted to validation.
            //////

            wrapper.setData({ keyLocal: { name: 'ShellHub', data: privateKey } });
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
});
