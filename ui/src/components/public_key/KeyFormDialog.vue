<template>
  <fragment>
    <v-tooltip
      v-if="createKey"
      bottom
      :disabled="hasAuthorization || action == 'private'"
    >
      <template #activator="{ on }">
        <div v-on="on">
          <v-btn
            :disabled="!hasAuthorization && action == 'public'"
            class="v-btn--active"
            text
            color="primary"
            data-test="createKey-btn"
          >
            Add {{ action }} Key
          </v-btn>
        </div>
      </template>

      <span>
        You don't have this kind of authorization.
      </span>
    </v-tooltip>

    <v-tooltip
      v-else
      :disabled="hasAuthorization"
      bottom
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-list-item-title
            data-test="close-item"
            v-on="on"
          >
            Edit
          </v-list-item-title>
        </span>

        <span v-on="on">
          <v-icon
            :disabled="!hasAuthorization"
            left
            data-test="remove-icon"
            v-on="on"
          >
            edit
          </v-icon>
        </span>
      </template>

      <span v-if="!hasAuthorization && action == 'public'">
        You don't have this kind of authorization.
      </span>
    </v-tooltip>

    <v-dialog
      v-model="showDialog"
      max-width="400"
      @click:outside="close"
    >
      <v-card data-test="keyFormDialog-card">
        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-title
            v-if="createKey"
            class="headline grey lighten-2 text-center"
          >
            New {{ action }} key
          </v-card-title>
          <v-card-title
            v-else
            class="headline grey lighten-2 text-center"
          >
            Edit {{ action }} key
          </v-card-title>

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
                :error-messages="errors"
                required
              />
            </ValidationProvider>

            <ValidationProvider
              v-if="action == 'public'"
              v-slot="{ errors }"
              name="Hostname"
            >
              <v-text-field
                v-model="keyLocal.hostname"
                label="Hostname"
                :error-messages="errors"
              />
            </ValidationProvider>

            <ValidationProvider
              v-if="action == 'public'"
              v-slot="{ errors }"
              name="Username"
              data-test="username-validationProvider"
            >
              <v-text-field
                v-model="keyLocal.username"
                label="Username"
                :error-messages="errors"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              ref="providerData"
              vid="key"
              name="Data"
              :rules="`required|parseKey:${action}`"
              :disabled="!createKey"
            >
              <v-textarea
                v-model="keyLocal.data"
                label="Data"
                :error-messages="errors"
                required
                :disabled="!createKey"
                :messages="supportedKeys"
              />
            </ValidationProvider>
          </v-card-text>

          <v-card-actions>
            <v-spacer />

            <v-btn
              text
              data-test="cancel-btn"
              @click="close"
            >
              Cancel
            </v-btn>

            <v-btn
              v-if="createKey"
              text
              data-test="create-btn"
              @click="passes(create)"
            >
              Create
            </v-btn>

            <v-btn
              v-else
              text
              data-test="edit-btn"
              @click="passes(edit)"
            >
              Edit
            </v-btn>
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
  name: 'KeyFormDialogComponent',

  filters: { hasPermission },

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

    createKey: {
      type: Boolean,
      required: true,
    },

    action: {
      type: String,
      default: 'public',
      required: false,
      validator: (value) => ['public', 'private'].includes(value),
    },

    show: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      keyLocal: [],
      supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
    };
  },

  computed: {
    showDialog: {
      get() {
        return this.show && this.hasAuthorization;
      },
      set(value) {
        this.$emit('update:show', value);
      },
    },

    hasAuthorization() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        let action = '';
        if (this.createKey) action = 'create';
        else action = 'edit';

        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.publicKey[action],
        );
      }

      return false;
    },
  },

  async created() {
    await this.setLocalVariable();
  },

  async updated() {
    await this.setLocalVariable();
  },

  methods: {
    setLocalVariable() {
      if (this.createKey) {
        this.keyLocal = {
          name: '',
          data: '',
        };

        if (this.action === 'public') this.keyLocal.hostname = '';
      } else {
        this.keyLocal = { ...this.keyObject };
        if (this.action === 'public') this.keyLocal.data = atob(this.keyObject.data);
      }
    },

    async create() {
      const keySend = this.keyLocal;

      switch (this.action) {
      case 'public':
        try {
          keySend.data = btoa(this.keyLocal.data);
          await this.$store.dispatch('publickeys/post', keySend);
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.publicKeyCreating);
          this.update();
        } catch (error) {
          if (error.response.status === 409) {
            this.$refs.obs.setErrors({
              key: error.response.data.message,
            });
          } else {
            this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.publicKeyCreating);
          }
        }
        break;
      case 'private':
        try {
          await this.$store.dispatch('privatekeys/set', keySend);
          this.$store.dispatch('snackbar/showSnackbarSuccessNotRequest', this.$success.privateKeyCreating);
          this.update();
        } catch (error) {
          switch (true) {
          case error.message === 'both': {
            this.$refs.obs.setErrors({
              name: ['The name already exists'],
              key: ['The private key data already exists'],
            });
            break;
          }
          case error.message === 'name': {
            this.$refs.obs.setErrors({
              name: ['The name already exists'],
            });
            break;
          }
          case error.message === 'private_key': {
            this.$refs.obs.setErrors({
              key: ['The private key data already exists'],
            });
            break;
          }
          default: {
            this.$store.dispatch('snackbar/showSnackbarErrorNotRequest', this.$errors.snackbar.privateKeyCreating);
          }
          }
        }
        break;
      default:
      }
    },

    async edit() {
      switch (this.action) {
      case 'public':
        try {
          await this.$store.dispatch('publickeys/put', this.keyLocal);
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.publicKeyEditing);
          this.update();
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.publicKeyEditing);
        }
        break;
      case 'private':
        try {
          await this.$store.dispatch('privatekeys/edit', this.keyLocal);
          this.$store.dispatch('snackbar/showSnackbarSuccessNotRequest', this.$success.privateKeyEditing);
          this.update();
        } catch (error) {
          if (error.message === 'name') {
            this.$refs.obs.setErrors({
              name: ['The name already exists'],
            });
          } else {
            this.$store.dispatch('snackbar/showSnackbarErrorNotRequest', this.$errors.snackbar.privateKeyEditing);
          }
        }
        break;
      default:
      }
    },

    update() {
      this.$emit('update');
      this.close();
    },

    close() {
      this.$emit('update:show', false);
      this.$refs.obs.reset();
    },
  },
};
</script>
