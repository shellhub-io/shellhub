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
      <v-card data-test="privateKeyFormDialog-card">
        <v-card-title
          class="headline primary"
          data-test="text-title"
          v-text="'Edit Private Key'"
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
              :disabled="true"
            >
              <v-textarea
                v-model="privateKey.data"
                label="Data"
                :error-messages="errors"
                required
                :messages="supportedKeys"
                :disabled="true"
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
  name: 'PublicKeyFormDialogAdd',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  props: {
    keyObject: {
      type: Object,
      required: true,
      default: Object,
    },

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
      this.privateKey = { ...this.keyObject };
    },

    async edit() {
      try {
        await this.$store.dispatch('privatekeys/edit', this.privateKey);
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
