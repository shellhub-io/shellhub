<template>
  <fragment>
    <v-btn
      color="primary"
      data-test="createKey-btn"
      v-text="'Add Private Key'"
    />

    <v-dialog
      v-model="showDialog"
      max-width="400"
      @click:outside="close"
    >
      <v-card data-test="privateKeyFormDialog-card">
        <v-card-title
          class="headline primary"
          data-test="text-title"
          v-text="'New Private Key'"
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
                v-model="privateKey.name"
                label="Name"
                :error-messages="errors"
                required
                data-test="name-field"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              ref="providerData"
              vid="key"
              name="Data"
              :rules="'required|parseKey:private'"
            >
              <v-textarea
                v-model="privateKey.data"
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

export default {
  name: 'PublicKeyFormDialogAdd',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  props: {
    show: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      privateKey: {
        name: '',
        data: '',
      },
      supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
    };
  },

  computed: {
    showDialog: {
      get() {
        return this.show;
      },

      set(value) {
        this.$emit('update:show', value);
      },
    },
  },

  async updated() {
    await this.setLocalVariable();
  },

  methods: {
    setLocalVariable() {
      this.privateKey = {
        name: '',
        data: '',
      };
    },

    async create() {
      try {
        await this.$store.dispatch('privatekeys/set', this.privateKey);
        this.$store.dispatch('snackbar/showSnackbarSuccessNotRequest', this.$success.privateKeyCreating);
        this.close();
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
    },

    close() {
      this.$emit('update:show', false);
      this.$refs.obs.reset();
    },
  },
};
</script>
