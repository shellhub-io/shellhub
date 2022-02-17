<template>
  <fragment>
    <v-list-item-icon class="mr-0">
      <v-icon
        left
        data-test="rename-icon"
        v-text="'mdi-pencil'"
      />
    </v-list-item-icon>

    <v-list-item-content>
      <v-list-item-title
        class="text-left"
        data-test="rename-title"
        v-text="'Rename'"
      />
    </v-list-item-content>

    <v-dialog
      v-model="showDialog"
      max-width="450"
      @click:outside="close"
    >
      <v-card data-test="deviceRename-card">
        <v-card-title
          class="headline primary"
          data-test="text-title"
          v-text="'Rename Device'"
        />

        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-text class="caption mb-0">
            <ValidationProvider
              v-slot="{ errors }"
              ref="providerHostname"
              name="Hostname"
              rules="required|rfc1123|noDot|device"
              vid="hostname"
            >
              <v-text-field
                v-model="editName"
                label="Hostname"
                :error-messages="errors"
                require
                :messages="messages"
                data-test="hostname-field"
              />
            </ValidationProvider>
          </v-card-text>

          <v-card-actions>
            <v-spacer />

            <v-btn
              text
              data-test="close-btn"
              @click="close()"
              v-text="'Close'"
            />

            <v-btn
              color="primary"
              text
              data-test="rename-btn"
              @click="passes(edit)"
              v-text="'Rename'"
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
  name: 'DeviceRenameComponent',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  props: {
    name: {
      type: String,
      required: true,
    },

    uid: {
      type: String,
      required: true,
    },

    show: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      invalid: false,
      editName: '',
      messages: 'Examples: (foobar, foo-bar-ba-z-qux, foo-example, 127-0-0-1)',
    };
  },

  computed: {
    device: {
      get() {
        return {
          name: this.name,
          uid: this.uid,
        };
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

  created() {
    this.editName = this.device.name;
  },

  updated() {
    this.editName = this.device.name;
  },

  methods: {
    cancel() {
      this.dialog = false;
      this.invalid = false;
      this.editName = '';
    },

    async edit() {
      try {
        await this.$store.dispatch('devices/rename', {
          uid: this.device.uid,
          name: this.editName,
        });

        this.close();
        this.$emit('new-hostname', this.editName);
        this.editName = '';
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.deviceRename);
      } catch (error) {
        if (error.response.status === 400) {
          this.$refs.obs.setErrors({
            hostname: this.$errors.form.invalid('hostname', 'nonStandardCharacters'),
          });
        } else if (error.response.status === 409) {
          this.$refs.obs.setErrors({
            hostname: ['The name already exists in the namespace'],
          });
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceRename);
        }
      }
    },

    close() {
      this.$emit('update:show', false);
    },
  },
};

</script>
