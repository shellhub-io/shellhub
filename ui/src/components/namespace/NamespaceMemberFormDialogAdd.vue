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
            class="mr-2"
            color="primary"
            data-test="addMember-btn"
            @click="dialog = !dialog"
          >
            Add Member
          </v-btn>
        </div>
      </template>

      <span>
        You don't have this kind of authorization.
      </span>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="450"
      @click:outside="close()"
    >
      <v-card data-test="namespaceNewMember-dialog">
        <v-card-title class="headline primary">
          Add member to namespace
        </v-card-title>

        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-text class="caption mb-0">
            <ValidationProvider
              v-slot="{ errors }"
              ref="providerUsername"
              name="Username"
              rules="required"
              vid="username"
            >
              <v-text-field
                v-model="member.username"
                label="Username"
                :error-messages="errors"
                require
                data-test="username-text"
              />
            </ValidationProvider>
          </v-card-text>

          <v-card-text class="caption mb-0 pt-0">
            <ValidationProvider
              v-slot="{ errors }"
              ref="providerRole"
              vid="role"
              name="role"
              rules="required"
            >
              <v-row align="center">
                <v-col cols="12">
                  <v-select
                    v-model="member.selectedRole"
                    :items="items"
                    label="Role"
                    :error-messages="errors"
                    require
                    data-test="role-select"
                  />
                </v-col>
              </v-row>
            </ValidationProvider>
          </v-card-text>

          <v-card-actions>
            <v-spacer />
            <v-btn
              text
              data-test="close-btn"
              @click="close()"
            >
              Close
            </v-btn>

            <v-btn
              color="primary"
              text
              data-test="add-btn"
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

import hasPermission from '@/components/filter/permission';

export default {
  name: 'NamespaceNewMemberFormDialogAdd',

  filters: { hasPermission },

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  data() {
    return {
      username: '',
      selectedRole: '',
      action: 'addMember',
      dialog: false,
      member: {
        username: '',
        selectedRole: '',
      },
      items: ['administrator', 'operator', 'observer'],
    };
  },

  computed: {
    hasAuthorization() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.namespace[this.action],
        );
      }

      return false;
    },
  },

  updated() {
    this.resetMemberVariable();
  },

  methods: {
    resetMemberVariable() {
      this.member.username = '';
      this.member.selectedRole = '';
    },

    async addMember() {
      try {
        await this.$store.dispatch('namespaces/addUser', {
          username: this.member.username,
          tenant_id: this.$store.getters['auth/tenant'],
          role: this.member.selectedRole,
        });

        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceNewMember);
        this.update();
      } catch (error) {
        if (error.response.status === 404) {
          this.$refs.obs.setErrors({
            username: 'The username doesn\'t exist.',
          });
        } else if (error.response.status === 409) {
          this.$refs.obs.setErrors({
            username: 'The username has already been added to namespace.',
          });
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.namespaceNewMember);
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
