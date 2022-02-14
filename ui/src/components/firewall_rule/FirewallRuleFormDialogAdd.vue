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
            data-test="add-btn"
            @click="dialog = !dialog"
            v-text="'Add Rule'"
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
      <v-card data-test="firewallRuleForm-card">
        <v-card-title
          class="headline primary"
          data-test="text-title"
          v-text="'New Rule'"
        />

        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-text>
            <v-layout
              justify-space-between
              align-center
            >
              <v-flex>
                <v-card :elevation="0">
                  <v-card-text
                    class="v-label theme--light pl-0"
                    v-text="'Active'"
                  />
                </v-card>
              </v-flex>

              <v-flex xs2>
                <v-card :elevation="0">
                  <v-switch v-model="ruleFirewall.active" />
                </v-card>
              </v-flex>
            </v-layout>

            <ValidationProvider
              v-slot="{ errors }"
              name="Priority"
              rules="required|integer"
            >
              <v-text-field
                v-model="ruleFirewall.priority"
                label="Priority"
                type="number"
                :error-messages="errors"
                required
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              name="Action"
              rules="required"
            >
              <v-select
                v-model="ruleFirewall.action"
                :items="state"
                item-text="name"
                item-value="id"
                label="Action"
                :error-messages="errors"
                required
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              name="Source IP"
              rules="required"
            >
              <v-text-field
                v-model="ruleFirewall.source_ip"
                label="Source IP"
                :error-messages="errors"
                required
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              name="Username"
              rules="required"
            >
              <v-text-field
                v-model="ruleFirewall.username"
                label="Username"
                :error-messages="errors"
                required
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              name="Hostname"
              rules="required"
            >
              <v-text-field
                v-model="ruleFirewall.hostname"
                label="Hostname"
                :error-messages="errors"
                required
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
  name: 'FirewallRuleFormDialogAdd',

  filters: { hasPermission },

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  data() {
    return {
      dialog: false,
      action: 'create',
      ruleFirewall: {
        active: true,
        priority: '',
        action: '',
        source_ip: '',
        username: '',
        hostname: '',
      },
      state: [
        {
          id: 'allow',
          name: 'allow',
        },
        {
          id: 'deny',
          name: 'deny',
        },
      ],
    };
  },

  computed: {
    hasAuthorization() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.firewall[this.action],
        );
      }

      return false;
    },
  },

  async updated() {
    await this.resetRuleFirewall();
  },

  methods: {
    resetRuleFirewall() {
      this.ruleFirewall.active = true;
      this.ruleFirewall.priority = '';
      this.ruleFirewall.action = '';
      this.ruleFirewall.source_ip = '';
      this.ruleFirewall.username = '';
      this.ruleFirewall.hostname = '';
    },

    async create() {
      try {
        await this.$store.dispatch('firewallrules/post', this.ruleFirewall);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.firewallRuleCreating);
        this.update();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.firewallRuleCreating);
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
