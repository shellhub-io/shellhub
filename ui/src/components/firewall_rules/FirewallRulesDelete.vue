<template>
  <fragment>
    <v-tooltip bottom>
      <template #activator="{ on }">
        <v-icon
          v-on="on"
          @click="dialog = !dialog"
        >
          delete
        </v-icon>
      </template>
      <span>Remove</span>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card>
        <v-card-title class="headline grey lighten-2 text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          You are about to remove this firewall rule
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            @click="dialog=!dialog"
          >
            Close
          </v-btn>

          <v-btn
            color="red darken-1"
            text
            @click="remove();"
          >
            Remove
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

export default {
  name: 'FirewallDelete',

  props: {
    id: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      dialog: false,
    };
  },

  methods: {
    async remove() {
      await this.$store.dispatch('firewallrules/remove', this.id);
      this.dialog = false;
    },
  },
};
</script>
