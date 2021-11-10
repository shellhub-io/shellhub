<template>
  <fragment>
    <v-tooltip bottom>
      <template #activator="{ on }">
        <span v-on="on">
          <v-icon
            :disabled="!hasAuthorization"
            v-on="on"
            @click="dialog = !dialog"
          >
            mdi-pencil
          </v-icon>
        </span>
      </template>

      <div>
        <span
          v-if="hasAuthorization"
          data-test="text-tooltip"
        >
          Edit
        </span>

        <span v-else>
          You don't have this kind of authorization.
        </span>
      </div>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="450"
      @click:outside="cancel"
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
              @click="cancel"
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
  },

  data() {
    return {
      dialog: false,
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

    hasAuthorization() {
      const accessType = this.$store.getters['auth/accessType'];
      if (accessType !== '') {
        return hasPermission(
          this.$authorizer.accessType[accessType],
          this.$actions.device.rename,
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
        this.dialog = false;
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
  },
};

</script>
