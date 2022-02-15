import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import PublicKeyList from '@/components/public_key/PublicKeyList';
import { actions, authorizer } from '../../../../src/authorizer';

describe('PublicKeyList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const statusGlobal = true;
  const numberPublicKeysGlobal = 2;

  const publicKeysGlobal = [
    {
      data: 'BBGVvbmF',
      fingerprint: '00:00:00',
      created_at: '2020-11-23T20:59:13.323Z',
      tenant_id: 'xxxxxxxx',
      name: 'shellhub',
    },
    {
      data: 'AbGVvbmF',
      fingerprint: '00:00:00',
      created_at: '2020-11-23T20:59:13.323Z',
      tenant_id: 'xxxxxxxx',
      name: 'shellhub',
    },
  ];

  const headers = [
    {
      text: 'Name',
      value: 'name',
      align: 'center',
    },
    {
      text: 'Fingerprint',
      value: 'fingerprint',
      align: 'center',
    },
    {
      text: 'Hostname',
      value: 'hostname',
      align: 'center',
    },
    {
      text: 'Username',
      value: 'username',
      align: 'center',
    },
    {
      text: 'Created At',
      value: 'created_at',
      align: 'center',
    },
    {
      text: 'Actions',
      value: 'actions',
      align: 'center',
    },
  ];

  const tests = [
    {
      description: 'List data when user has owner role',
      role: {
        type: 'owner',
        permission: true,
      },
      variables: {
        publicKeysGlobal,
        numberPublicKeysGlobal,
        statusGlobal,
      },
      data: {
        pagination: {},
        publicKeyFormDialogShow: [],
        publicKeyDeleteShow: [],
        editAction: 'edit',
        headers,
      },
      computed: {
        getPublicKeys: publicKeysGlobal,
        getNumberPublicKeys: numberPublicKeysGlobal,
        hasAuthorizationFormDialogEdit: true,
        hasAuthorizationFormDialogRemove: true,
      },
    },
    {
      description: 'List data when user has operator role',
      role: {
        type: 'operator',
        permission: false,
      },
      variables: {
        publicKeysGlobal,
        numberPublicKeysGlobal,
        statusGlobal,
      },
      data: {
        pagination: {},
        publicKeyFormDialogShow: [],
        publicKeyDeleteShow: [],
        editAction: 'edit',
        headers,
      },
      computed: {
        getPublicKeys: publicKeysGlobal,
        getNumberPublicKeys: numberPublicKeysGlobal,
        hasAuthorizationFormDialogEdit: false,
        hasAuthorizationFormDialogRemove: false,
      },
    },
  ];

  const storeVuex = (publicKeys, numberPublicKeys, status, currentRole) => new Vuex.Store({
    namespaced: true,
    state: {
      publicKeys,
      numberPublicKeys,
      status,
      currentRole,
    },
    getters: {
      'publickeys/list': (state) => state.publicKeys,
      'publickeys/getNumberPublicKeys': (state) => state.numberPublicKeys,
      'boxs/getStatus': (state) => state.status,
      'auth/role': (state) => state.currentRole,
    },
    actions: {
      'publickeys/fetch': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
      'boxs/setStatus': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = shallowMount(PublicKeyList, {
          store: storeVuex(
            test.variables.publicKeysGlobal,
            test.variables.numberPublicKeysGlobal,
            test.variables.statusGlobal,
            test.role.type,
          ),
          localVue,
          stubs: ['fragment'],
          vuetify,
          mocks: {
            $authorizer: authorizer,
            $actions: actions,
          },
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
      // Data and Props checking
      //////

      it('Compare data with default value', () => {
        Object.keys(test.data).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.data[item]);
        });
      });
      it('Process data in the computed', () => {
        Object.keys(test.computed).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.computed[item]);
        });
      });

      //////
      // HTML validation
      //////

      it('Renders the template with data', () => {
        const dt = wrapper.find('[data-test="publicKeyList-dataTable"]');
        const dataTableProps = dt.vm.$options.propsData;

        expect(dataTableProps.items).toHaveLength(test.variables.numberPublicKeysGlobal);
      });
    });
  });
});
