<template>
<v-card>
    <v-toolbar card color="transparent">
        <v-toolbar-title>Device Fleeta</v-toolbar-title>
    </v-toolbar>
    <v-divider></v-divider>
    <v-card-text class="pa-0">
        <v-list two-line class="pa-0">
            <template v-for="(item, index) in devices">
                <v-list-tile :key="item.uid">
                    <v-list-tile-avatar>
                        <v-icon color="success" v-if="item.online">
                            check_circle
                        </v-icon>
                        <v-icon v-else>check_circle</v-icon>
                    </v-list-tile-avatar>
                    <v-list-tile-content>
                        <v-list-tile-title v-html="item.uid"></v-list-tile-title>
                        <v-list-tile-sub-title v-html="item.identity.mac"></v-list-tile-sub-title>
                    </v-list-tile-content>
                    <v-list-tile-action>
                        <TerminalDialog :uid="item.uid" v-if="item.online"></TerminalDialog>
                    </v-list-tile-action>
                    <v-list-tile-action>
                        <v-list-tile-action-text v-if="!item.online">last seen {{ item.last_seen | moment("from", "now") }}</v-list-tile-action-text>
                    </v-list-tile-action>
                </v-list-tile>

                <v-divider v-if="index + 1 < devices.length" :key="index"></v-divider>
            </template>
        </v-list>
    </v-card-text>
</v-card>
</template>

<script>
import TerminalDialog from "@/components/TerminalDialog.vue";

export default {
  components: {
    TerminalDialog
  },

  created() {
    this.$store.dispatch("devices/fetch");
  },

  computed: {
    devices() {
      return this.$store.getters["devices/list"];
    }
  }
};
</script>
