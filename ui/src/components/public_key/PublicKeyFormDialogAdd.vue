<template>
  <fragment>
    <v-tooltip
      bottom
      :disabled="hasAuthorization"
    >
      <template #activator="{ on }">
        <div v-on="on">
          <v-btn
            :disabled="!hasAuthorization"
            color="primary"
            data-test="createKey-btn"
            @click="dialog = !dialog"
            v-text="'Add Public Key'"
          />
        </div>
      </template>

      <span>
        You don't have this kind of authorization.
      </span>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="400"
      @click:outside="close"
    >
      <v-card data-test="publicKeyFormDialog-card">
        <v-card-title
          class="headline primary"
          data-test="text-title"
          v-text="'New Public Key'"
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
                label="Name"
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
            >
              <v-textarea
                v-model="keyLocal.data"
                class="mt-5"
                label="Public key data"
                :error-messages="errors"
                required
                :messages="supportedKeys"
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
              data-test="create-btn"
              @click="passes(create)"
              v-text="'Create'"
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

import hasPermission from '@/components/filter/permission';

export default {
  name: 'PublickKeyFormDialogAdd',

  filters: { hasPermission },

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  data() {
    return {
      dialog: false,
      action: 'create',
      validateLength: true,
      username: '',
      choiceFilter: 'all',
      choiceUsername: 'all',
      tagChoices: [], // defaults to public key tags
      hostname: '',
      errMsg: '',
      keyLocal: {},
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
      supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
    };
  },

  computed: {
    tagNames() {
      return this.$store.getters['tags/list'];
    },

    hasAuthorization() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.publicKey[this.action],
        );
      }

      return false;
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

    dialog(val) {
      if (!val) {
        this.setLocalVariable();
      }
    },
  },

  methods: {
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

    setLocalVariable() {
      this.keyLocal = {};
      this.hostname = '';
      this.tagChoices = [];
      this.choiceFilter = 'all';
      this.choiceUsername = 'all';
    },

    async create() {
      try {
        this.chooseFilter();
        this.chooseUsername();
        const keySend = { ...this.keyLocal, data: btoa(this.keyLocal.data) };

        await this.$store.dispatch('publickeys/post', keySend);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.publicKeyCreating);
        this.update();
      } catch (error) {
        if (error.response.status === 409) {
          this.$refs.obs.setErrors({
            key: 'Public key data already exists',
          });
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.publicKeyCreating);
        }
      }
    },

    update() {
      this.$emit('update');
      this.close();
    },

    close() {
      this.dialog = false;
      this.$refs.obs.reset();
    },
  },
};
</script>
