<template>
  <v-form>
    <v-container>
      <v-row
        align="center"
        justify="center"
        class="mt-4"
      >
        <v-col
          sm="8"
        >
          <div
            class="mt-6 pl-4 pr-4"
          >
            <v-row>
              <v-col md="auto">
                <v-card
                  tile
                  :elevation="0"
                >
                  Tenant ID:
                </v-card>
              </v-col>
              <v-col
                md="auto"
                class="ml-auto"
              >
                <v-card
                  class="auto"
                  tile
                  :elevation="0"
                >
                  <v-chip>
                    <span>
                      {{ tenant }}
                    </span>
                    <v-icon
                      v-clipboard="tenant"
                      v-clipboard:success="() => {
                        this.$store.dispatch('modals/showSnackbarCopy', this.$copy.tenantId);
                      }"
                      right
                    >
                      mdi-content-copy
                    </v-icon>
                  </v-chip>
                </v-card>
              </v-col>
            </v-row>
          </div>

          <v-divider />
          <v-divider />

          <ValidationObserver
            ref="obs"
            v-slot="{ validated, passes }"
          >
            <div
              class="mt-6 pl-4 pr-4"
            >
              <ValidationProvider
                v-slot="{ errors }"
                name="Priority"
                rules="required"
              >
                <v-text-field
                  v-model="username"
                  label="Username"
                  :error-messages="errors"
                  required
                  :disabled="!editDataStatus"
                />
              </ValidationProvider>

              <ValidationProvider
                v-slot="{ errors }"
                name="Priority"
                rules="required|email"
              >
                <v-text-field
                  v-model="email"
                  class="mt-1 mb-4"
                  label="E-mail"
                  :error-messages="errors"
                  required
                  :disabled="!editDataStatus"
                />
              </ValidationProvider>

              <v-btn
                v-if="!editDataStatus"
                class="mr-2"
                outlined
                @click="editDataStatus = !editDataStatus"
              >
                Change Data
              </v-btn>

              <div
                v-if="editDataStatus"
              >
                <v-btn
                  class="mr-2 mt-4"
                  outlined
                  @click="editDataStatus = !editDataStatus"
                >
                  Cancel
                </v-btn>

                <v-btn
                  class="mr-2 mt-4"
                  outlined
                  @click="passes(updateData)"
                >
                  Save
                </v-btn>
              </div>
            </div>
          </ValidationObserver>

          <ValidationObserver
            ref="obs"
            v-slot="{ invalid, validated, passes }"
          >
            <v-divider class="mt-6" />
            <v-divider class="mb-6" />

            <div
              class="mt-6 pl-4 pr-4"
            >
              <ValidationProvider
                v-slot="{ errors }"
                name="Priority"
                rules="required"
                vid="currentPassword"
              >
                <v-text-field
                  v-model="currentPassword"
                  type="password"
                  label="Current password"
                  class="mb-4"
                  :error-messages="errors"
                  required
                  :disabled="!editPasswordStatus"
                />
              </ValidationProvider>

              <ValidationProvider
                v-slot="{ errors }"
                name="Priority"
                rules="required|password|comparePasswords:@currentPassword"
                vid="newPassword"
              >
                <v-text-field
                  v-model="newPassword"
                  type="password"
                  label="New password"
                  class="mb-4"
                  :error-messages="errors"
                  required
                  :disabled="!editPasswordStatus"
                />
              </ValidationProvider>

              <ValidationProvider
                v-slot="{ errors }"
                rules="required|confirmed:newPassword"
                name="confirm"
              >
                <v-text-field
                  v-model="newPasswordConfirm"
                  label="Confirm new password"
                  type="password"
                  class="mb-4"
                  :error-messages="errors"
                  required
                  :disabled="!editPasswordStatus"
                />
              </ValidationProvider>

              <v-btn
                v-if="!editPasswordStatus"
                class="mr-2"
                outlined
                @click="editPasswordStatus = !editPasswordStatus"
              >
                Change Password
              </v-btn>

              <div
                v-if="editPasswordStatus"
              >
                <v-btn
                  class="mr-2"
                  outlined
                  @click="editPasswordStatus = !editPasswordStatus"
                >
                  Cancel
                </v-btn>

                <v-btn
                  class="mr-2"
                  outlined
                  @click="passes(updatePassword)"
                >
                  Save
                </v-btn>
              </div>
            </div>
          </ValidationObserver>
        </v-col>
      </v-row>
    </v-container>
  </v-form>
</template>

<script>

import {
  ValidationObserver,
  ValidationProvider,
} from 'vee-validate';

export default {
  name: 'SettingProfile',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  data() {
    return {
      username: '',
      email: '',
      currentPassword: '',
      newPassword: '',
      newPasswordConfirm: '',
      editDataStatus: false,
      editPasswordStatus: false,
      show: false,
    };
  },

  computed: {
    tenant() {
      return this.$store.getters['auth/tenant'];
    },
  },

  created() {
    this.setData();
  },

  methods: {
    setData() {
      this.username = this.$store.getters['auth/currentUser'];
      this.email = this.$store.getters['auth/email'];
    },

    enableEdit() {
      if (this.editDataStatus) {
        this.editDataStatus = !this.editDataStatus;
      } else if (this.editPasswordStatus) {
        this.editPasswordStatus = !this.editPasswordStatus;
      }
    },

    async updateData() {
      const data = {
        username: this.username,
        email: this.email,
      };

      try {
        await this.$store.dispatch('users/put', data);
        this.$store.dispatch('auth/changeUserData', data);
        this.$store.dispatch('modals/showSnackbarSuccessDefault');
        this.enableEdit();
      } catch {
        this.$store.dispatch('modals/showSnackbarErrorDefault');
      }
    },

    async updatePassword() {
      const data = {
        currentPassword: this.currentPassword,
        newPassword: this.newPassword,
      };

      try {
        await this.$store.dispatch('users/put', data);
        this.$store.dispatch('modals/showSnackbarSuccessDefault');
        this.enableEdit();
      } catch {
        this.$store.dispatch('modals/showSnackbarErrorDefault');
      }
    },
  },

};
</script>
