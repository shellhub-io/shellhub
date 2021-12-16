<template>
  <fragment>
    <v-tooltip
      :disabled="hasAuthorization"
      bottom
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-list-item-title
            data-test="rename-item"
            v-on="on"
          >
            Rename
          </v-list-item-title>
        </span>

        <span v-on="on">
          <v-icon
            :disabled="!hasAuthorization"
            left
            data-test="rename-icon"
            v-on="on"
          >
            mdi-pencil
          </v-icon>
        </span>
      </template>

      <span v-if="!hasAuthorization">
        You don't have this kind of authorization.
      </span>
    </v-tooltip>

    <v-dialog
      v-model="showDialog"
      max-width="450"
      @click:outside="close"
    >
      <v-card data-test="deviceRename-card">
        <v-card-title class="headline grey lighten-2 text-center">
          Rename Device
        </v-card-title>
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
              />
            </ValidationProvider>
          </v-card-text>

          <v-card-actions>
            <v-spacer />
            <v-btn
              text
              data-test="cancel-btn"
              @click="close()"
            >
              Close
            </v-btn>

            <v-btn
              color="primary"
              text
              data-test="rename-btn"
              @click="passes(edit)"
            >
              Rename
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
  name: 'DeviceRenameComponent',

  filters: { hasPermission },

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
      action: 'rename',
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
        return this.show && this.hasAuthorization;
      },
      set(value) {
        this.$emit('update:show', value);
      },
    },

    hasAuthorization() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.device[this.action],
        );
      }

      return false;
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
