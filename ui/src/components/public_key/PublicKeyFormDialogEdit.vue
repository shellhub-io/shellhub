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
      max-width="400"
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
                v-model="keyLocal.hostname"
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
                v-model="keyLocal.username"
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
              :rules="'required|parseKey:public'"
              :disabled="true"
            >
              <v-textarea
                v-model="keyLocal.data"
                label="Data"
                :error-messages="errors"
                required
                :disabled="true"
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
      required: true,
    },
  },

  data() {
    return {
      keyLocal: {
        name: '',
        hostname: '',
        username: '',
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

  async created() {
    await this.setLocalVariable();
  },

  async updated() {
    await this.setLocalVariable();
  },

  methods: {
    setLocalVariable() {
      this.keyLocal = { ...this.keyObject };
      this.keyLocal.data = atob(this.keyObject.data);
    },

    async edit() {
      try {
        await this.$store.dispatch('publickeys/put', this.keyLocal);
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
      this.$emit('update:show', false);
      this.$refs.obs.reset();
    },
  },
};
</script>
