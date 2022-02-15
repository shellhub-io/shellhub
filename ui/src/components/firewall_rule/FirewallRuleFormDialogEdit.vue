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
      <v-card data-test="firewallRuleForm-card">
        <v-card-title
          class="headline primary"
          data-test="text-title"
          v-text="'Edit Firewall Rule'"
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
                  <v-switch v-model="ruleFirewallLocal.active" />
                </v-card>
              </v-flex>
            </v-layout>

            <ValidationProvider
              v-slot="{ errors }"
              name="Priority"
              rules="required|integer"
            >
              <v-text-field
                v-model="ruleFirewallLocal.priority"
                label="Priority"
                type="number"
                :error-messages="errors"
                required
                data-test="priority-field"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              name="Action"
              rules="required"
            >
              <v-select
                v-model="ruleFirewallLocal.action"
                :items="state"
                item-text="name"
                item-value="id"
                label="Action"
                :error-messages="errors"
                required
                data-test="action-field"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              name="Source IP"
              rules="required"
            >
              <v-text-field
                v-model="ruleFirewallLocal.source_ip"
                label="Source IP"
                :error-messages="errors"
                required
                data-test="source_ip-field"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              name="Username"
              rules="required"
            >
              <v-text-field
                v-model="ruleFirewallLocal.username"
                label="Username"
                :error-messages="errors"
                required
                data-test="username-field"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              name="Hostname"
              rules="required"
            >
              <v-text-field
                v-model="ruleFirewallLocal.hostname"
                label="Hostname"
                :error-messages="errors"
                required
                data-test="hostname-field"
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
  name: 'FirewallRuleFormDialogComponent',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  props: {
    firewallRule: {
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
      dialog: false,
      state: [{
        id: 'allow',
        name: 'allow',
      },
      {
        id: 'deny',
        name: 'deny',
      }],

      ruleFirewallLocal: {
        active: true,
        priority: '',
        action: '',
        source_ip: '',
        username: '',
        hostname: '',
      },
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
      this.ruleFirewallLocal = { ...this.firewallRule };
    },

    async edit() {
      try {
        await this.$store.dispatch('firewallrules/put', this.ruleFirewallLocal);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.firewallRuleEditing);
        this.update();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.firewallRuleEditing);
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
