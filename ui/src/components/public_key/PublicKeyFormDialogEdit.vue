<template>
  <fragment>
    <v-list-item-icon class="mr-0">
      <v-icon
        left
        data-test="edit-icon"
        v-text="'edit'"
      />
    </v-list-item-icon>

    <v-list-item-content>
      <v-list-item-title
        class="text-left"
        data-test="edit-title"
        v-text="'Edit'"
      />
    </v-list-item-content>

    <v-dialog
      v-model="showDialog"
      max-width="520"
      @click:outside="close"
    >
      <v-card data-test="publicKeyFormDialog-card">
        <v-card-title
          class="headline primary"
          data-test="text-title"
          v-text="'Edit Public Key'"
        />

        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-text>
            <ValidationProvider
              v-slot="{ errors }"
              ref="providerName"
              vid="name"
              name="Name"
              rules="required"
            >
              <v-text-field
                v-model="keyLocal.name"
                label="Key name"
                placeholder="Name used to identify the public key"
                :error-messages="errors"
                required
                data-test="name-field"
              />
            </ValidationProvider>

            <v-row class="mt-2 mb-1 px-3">
              <v-select
                v-model="choiceUsername"
                label="Device username access restriction"
                :items="usernameList"
                item-text="filterText"
                item-value="filterName"
                data-test="access-restriction-field"
              />
            </v-row>

            <ValidationProvider
              v-if="choiceUsername==='username'"
              v-slot="{ errors }"
              name="Username"
              data-test="username-validationProvider"
            >
              <v-text-field
                v-model="username"
                label="Username"
                placeholder="Username used during the connection"
                :error-messages="errors"
                data-test="username-field"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              name="Hostname"
            >
              <v-row class="mt-1 px-3">
                <v-select
                  v-model="choiceFilter"
                  label="Device access restriction"
                  :items="filterList"
                  item-text="filterText"
                  item-value="filterName"
                  data-test="access-restriction-field"
                />
              </v-row>

              <v-row class="px-3">
                <v-select
                  v-if="choiceFilter === 'tags'"
                  v-model="tagChoices"
                  :items="tagNames"
                  data-test="tags-field"
                  attach
                  chips
                  label="Tags"
                  :rules="[validateLength]"
                  :error-messages="errMsg"
                  :menu-props="{ top: true, maxHeight: 150, offsetY: true }"
                  multiple
                />
                <v-text-field
                  v-if="choiceFilter === 'hostname'"
                  v-model="hostname"
                  label="Hostname"
                  :error-messages="errors"
                  data-test="hostname-field"
                />
              </v-row>
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              ref="providerData"
              vid="key"
              name="Data"
              :rules="'required|parseKey:public'"
              :disabled="true"
            >
              <v-textarea
                v-model="keyLocal.data"
                class="mt-5"
                label="Public key data"
                :error-messages="errors"
                required
                :disabled="true"
                :messages="supportedKeys"
                placeholder="Data"
                data-test="data-field"
                rows="2"
              />
            </ValidationProvider>
          </v-card-text>

          <v-card-actions>
            <v-spacer />

            <v-btn
              text
              data-test="cancel-btn"
              @click="close"
              v-text="'Cancel'"
            />

            <v-btn
              text
              data-test="edit-btn"
              @click="passes(edit)"
              v-text="'Edit'"
            />
          </v-card-actions>
        </ValidationObserver>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

import {
  ValidationObserver,
  ValidationProvider,
} from 'vee-validate';

export default {
  name: 'PublicKeyFormDialogEdit',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  props: {
    keyObject: {
      type: Object,
      required: false,
      default: Object,
    },

    show: {
      type: Boolean,
      required: false,
    },
  },

  data() {
    return {
      choiceFilter: 'hostname',
      dialog: false,
      validateLength: true,
      username: '',
      errMsg: '',
      choiceUsername: 'username',
      filterList: [
        {
          filterName: 'all',
          filterText: 'Allow the key to connect to all available devices',
        },
        {
          filterName: 'hostname',
          filterText: 'Restrict access using a regexp for hostname',
        },
        {
          filterName: 'tags',
          filterText: 'Restrict access by tags',
        },
      ],
      usernameList: [
        {
          filterName: 'all',
          filterText: 'Allow any user',
        },
        {
          filterName: 'username',
          filterText: 'Restrict access using a regexp for username',
        },
      ],
      tagChoices: [],
      hostname: '',
      keyLocal: {
        name: '',
        username: '',
        data: '',
      },
      supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
    };
  },

  computed: {
    hasTags() {
      const { keyObject } = this.$props;
      if (!keyObject) return false;
      return Reflect.ownKeys(keyObject.filter)[0] === 'tags';
    },

    tagNames: {
      get() {
        return this.$store.getters['tags/list'];
      },

      set(val) {
        this.tagChoices = val;
      },
    },

    showDialog: {
      get() {
        return this.show;
      },

      set(value) {
        this.$emit('update:show', value);
      },
    },
  },

  watch: {
    tagChoices(list) {
      if (list.length > 3) {
        this.validateLength = false;
        this.$nextTick(() => this.tagChoices.pop());
        this.errMsg = 'The maximum capacity has reached';
      } else if (list.length <= 2) {
        this.validateLength = true;
        this.errMsg = '';
      }
    },
  },

  async created() {
    await this.setLocalVariable();
  },

  async updated() {
    this.handleUpdate();
    await this.setLocalVariable();
  },

  methods: {
    handleUpdate() {
      if (this.showDialog) {
        if (this.hasTags) {
          const { tags } = this.$props.keyObject.filter;
          this.tagChoices = tags;
          this.choiceFilter = 'tags';
        } else {
          const { hostname } = this.$props.keyObject.filter;
          if (!!hostname && hostname !== '.*') {
            this.choiceFilter = 'hostname';
            this.hostname = hostname;
          }
        }

        const { username } = this.$props.keyObject;
        this.choiceUsername = (username === '' ? 'all' : 'username');
        this.username = username;
      }
    },

    chooseFilter() {
      switch (this.choiceFilter) {
      case 'all': {
        this.keyLocal = { ...this.keyLocal, filter: { hostname: '.*' } };
        break;
      }
      case 'hostname': {
        this.keyLocal = { ...this.keyLocal, filter: { hostname: this.hostname } };
        break;
      }
      case 'tags': {
        this.keyLocal = { ...this.keyLocal, filter: { tags: this.tagChoices } };
        break;
      }
      default:
      }
    },

    chooseUsername() {
      switch (this.choiceUsername) {
      case 'all': {
        this.keyLocal = { ...this.keyLocal, username: '' };
        break;
      }
      case 'username': {
        this.keyLocal = { ...this.keyLocal, username: this.username };
        break;
      }
      default:
      }
    },

    setLocalVariable() {
      this.keyLocal = { ...this.keyObject };
      this.keyLocal.data = atob(this.keyObject.data);
    },

    async edit() {
      let keySend = this.publicKey;
      this.chooseFilter();
      this.chooseUsername();
      keySend = { ...this.keyLocal, data: btoa(this.keyLocal.data) };

      try {
        await this.$store.dispatch('publickeys/put', keySend);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.publicKeyEditing);
        this.update();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.publicKeyEditing);
      }
    },

    update() {
      this.$emit('update');
      this.close();
    },

    close() {
      this.hostname = '';
      this.tagChoices = [];
      this.$emit('update:show', false);
      this.$refs.obs.reset();
    },
  },
};
</script>
