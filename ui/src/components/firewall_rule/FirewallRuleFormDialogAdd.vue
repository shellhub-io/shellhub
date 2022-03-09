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
      max-width="520"
      @click:outside="close"
    >
      <v-card data-test="firewallRuleForm-card">
        <v-card-title
          class="headline primary"
          data-test="text-title"
          v-text="'New Firewall Rule'"
        />

        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-text>
            <v-row>
              <v-col>
                <v-select
                  v-model="ruleFirewall.status"
                  :items="ruleStatus"
                  item-text="text"
                  item-value="type"
                  label="Rule status"
                  required
                />
              </v-col>

              <v-col>
                <ValidationProvider
                  v-slot="{ errors }"
                  name="Priority"
                  rules="required|integer"
                >
                  <v-text-field
                    v-model="ruleFirewall.priority"
                    label="Rule priority"
                    type="number"
                    :error-messages="errors"
                    required
                  />
                </ValidationProvider>
              </v-col>

              <v-col>
                <ValidationProvider
                  v-slot="{ errors }"
                  name="Action"
                  rules="required"
                >
                  <v-select
                    v-model="ruleFirewall.policy"
                    :items="state"
                    item-text="name"
                    item-value="id"
                    label="Rule policy"
                    :error-messages="errors"
                    required
                  />
                </ValidationProvider>
              </v-col>
            </v-row>

            <v-row class="mt-1 mb-1 px-3">
              <v-select
                v-model="choiceIP"
                label="Source IP access restriction"
                :items="sourceIPFieldChoices"
                item-text="filterText"
                item-value="filterName"
                data-test="source_ip-field"
              />
            </v-row>

            <ValidationProvider
              v-if="choiceIP==='ipDetails'"
              v-slot="{ errors }"
              name="Source IP"
              rules="required"
            >
              <v-text-field
                v-model="ruleFirewall.source_ip"
                label="Rule source IP"
                :error-messages="errors"
                required
              />
            </ValidationProvider>

            <v-row class="mt-1 mb-1 px-3">
              <v-select
                v-model="choiceUsername"
                label="Device username access restriction"
                :items="usernameFieldChoices"
                item-text="filterText"
                item-value="filterName"
                data-test="username-field"
              />
            </v-row>

            <ValidationProvider
              v-if="choiceUsername==='username'"
              v-slot="{ errors }"
              name="Username"
              rules="required"
            >
              <v-text-field
                v-model="ruleFirewall.username"
                label="Username access restriction"
                placeholder="Username used during the connection"
                :error-messages="errors"
                required
              />
            </ValidationProvider>

            <v-row class="mt-2 mb-1 px-3">
              <v-select
                v-model="choiceFilter"
                label="Device access restriction"
                :items="filterFieldChoices"
                item-text="filterText"
                item-value="filterName"
                data-test="device-field"
              />
            </v-row>

            <ValidationProvider
              v-if="choiceFilter==='hostname'"
              v-slot="{ errors }"
              name="Hostname"
              rules="required"
            >
              <v-text-field
                v-model="filterField"
                label="Device hostname access restriction"
                placeholder="Device hostname used during the connection"
                :error-messages="errors"
                required
                data-test="hostname-field"
              />
            </ValidationProvider>

            <v-row
              v-if="choiceFilter==='tags'"
              class="px-3 mt-2"
            >
              <v-select
                v-model="tagChoices"
                :items="tagNames"
                data-test="tags-selector"
                attach
                chips
                label="Tags"
                :rules="[validateLength]"
                :error-messages="errMsg"
                :menu-props="{ top: true, maxHeight: 150, offsetY: true }"
                multiple
              />
            </v-row>
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
      choiceUsername: 'all',
      choiceIP: 'all',
      choiceFilter: 'all',
      validateLength: true,
      filterField: '',
      errMsg: '',
      ruleStatus: [
        {
          type: 'active',
          text: 'Active',
        },
        {
          type: 'inactive',
          text: 'Inactive',
        },
      ],
      tagChoices: [],
      sourceIPFieldChoices: [
        {
          filterName: 'all',
          filterText: 'Define source IP to all devices',
        },
        {
          filterName: 'ipDetails',
          filterText: 'Restrict source IP through a regexp',
        },
      ],
      filterFieldChoices: [
        {
          filterName: 'all',
          filterText: 'Define rule to all devices',
        },
        {
          filterName: 'hostname',
          filterText: 'Restrict rule with a regexp for hostname',
        },
        {
          filterName: 'tags',
          filterText: 'Restrict rule by device tags',
        },
      ],
      usernameFieldChoices: [
        {
          filterName: 'all',
          filterText: 'Define rule to all users',
        },
        {
          filterName: 'username',
          filterText: 'Restrict access using a regexp for username',
        },
      ],
      ruleFirewall: {
        policy: 'allow',
        priority: '',
        status: 'active',
        source_ip: '',
        username: '',
      },
      state: [
        {
          id: 'allow',
          name: 'Allow',
        },
        {
          id: 'deny',
          name: 'Deny',
        },
      ],
    };
  },

  computed: {
    tagNames() {
      return this.$store.getters['tags/list'];
    },

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

  watch: {
    tagChoices(list) {
      if (list.length > 3) {
        this.validateLength = false;
        this.$nextTick(() => this.tagChoices.pop());
        this.errMsg = 'The maximum capacity has reached';
      } else if (list.length === 0) {
        this.validateLength = false;
        this.errMsg = 'You must choose at least one tag';
      } else if (list.length <= 2) {
        this.validateLength = true;
        this.errMsg = '';
      }
    },
  },

  async updated() {
    await this.resetRuleFirewall();
  },

  methods: {
    resetRuleFirewall() {
      this.ruleFirewall = {
        policy: 'allow',
        priority: '',
        status: 'active',
        source_ip: '',
        username: '',
      };
      this.choiceFilter = 'all';
      this.choiceUsername = 'all';
      this.choiceIP = 'all';
      this.tagChoices = [];
      this.validateLength = true;
      this.errMsg = '';
    },

    constructFilterObject() {
      let filterObj = {};

      if (this.choiceFilter === 'hostname') {
        filterObj = { hostname: this.filterField };
      } else if (this.choiceFilter === 'tags') {
        filterObj = { tags: this.tagChoices };
      } else if (this.choiceFilter === 'all') {
        filterObj = { hostname: '.*' };
      }

      if (this.choiceUsername === 'all') {
        this.ruleFirewall.username = '.*';
      }

      if (this.choiceIP === 'all') {
        this.ruleFirewall.source_ip = '.*';
      }

      this.ruleFirewall = {
        ...this.ruleFirewall,
        filter: filterObj,
      };
    },

    async create() {
      this.constructFilterObject();

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
