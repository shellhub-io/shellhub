<template>
  <fragment>
    <v-btn
      class="ml-12"
      outlined
      @click="dialog = !dialog"
    >
      Add Member
    </v-btn>
    <v-dialog
      v-model="dialog"
      max-width="450"
      @click:outside="cancel"
    >
      <v-card>
        <v-card-title class="headline grey lighten-2 text-center">
          Add member to Namespace
        </v-card-title>
        <ValidationObserver
          ref="newuser"
          v-slot="{ passes }"
        >
          <v-card-text class="caption mb-0">
            <ValidationProvider
              v-slot="{ errors }"
              vid="username"
              name="user"
              rules="required"
            >
              <v-text-field
                v-model="username"
                label="Username"
                :error-messages="errors"
                require
              />
            </ValidationProvider>
          </v-card-text>
          <v-card-actions>
            <v-spacer />
            <v-btn
              text
              @click="cancel"
            >
              Close
            </v-btn>
            <v-btn
              color="primary"
              text
              @click="passes(addMember)"
            >
              Add
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

export default {
  name: 'NamespaceNewMember',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  props: {
    nsTenant: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      dialog: false,
      username: '',
    };
  },

  computed: {
    tenant() {
      return this.$props.nsTenant;
    },
  },

  methods: {
    cancel() {
      this.dialog = false;
      this.$refs.newuser.reset();
      this.username = '';
    },

    async addMember() {
      try {
        await this.$store.dispatch('namespaces/addUser', {
          username: this.username,
          tenant_id: this.tenant,
        });
        await this.$store.dispatch('namespaces/get', this.tenant);
        this.dialog = false;
        this.username = '';
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceNewMember);
      } catch (error) {
        if (error.response.status === 404) {
          this.$refs.newuser.setErrors({
            username: 'The username doesn\'t exist.',
          });
        } else if (error.response.status === 409) {
          this.$refs.newuser.setErrors({
            username: 'The username has already been added to namespace.',
          });
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.namespaceNewMember);
        }
      }
    },
  },
};

</script>
