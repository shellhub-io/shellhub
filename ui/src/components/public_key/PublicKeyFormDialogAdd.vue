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
                v-model="publicKey.name"
                label="Name"
                :error-messages="errors"
                required
                data-test="name-field"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              name="Hostname"
            >
              <v-text-field
                v-model="publicKey.hostname"
                label="Hostname"
                :error-messages="errors"
                data-test="hostname-field"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              name="Username"
              data-test="username-validationProvider"
            >
              <v-text-field
                v-model="publicKey.username"
                label="Username"
                :error-messages="errors"
                data-test="username-field"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              ref="providerData"
              vid="key"
              name="Data"
              :rules="`required|parseKey:${action}`"
            >
              <v-textarea
                v-model="publicKey.data"
                label="Data"
                :error-messages="errors"
                required
                :messages="supportedKeys"
                data-test="data-field"
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
      publicKey: {
        name: '',
        hostname: '',
        username: '',
        data: '',
      },
      supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
    };
  },

  computed: {
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

  async updated() {
    await this.setLocalVariable();
  },

  methods: {
    setLocalVariable() {
      this.publicKey.name = '';
      this.publicKey.hostname = '';
      this.publicKey.username = '';
      this.publicKey.data = '';
    },

    async create() {
      try {
        const keySend = this.publicKey;
        keySend.data = btoa(this.publicKey.data);

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
