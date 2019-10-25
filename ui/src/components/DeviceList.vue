<template>
<v-card>
    <v-app-bar flat color="transparent">
        <v-toolbar-title>Device Fleet</v-toolbar-title>
    </v-app-bar>
    <v-divider></v-divider>
    <v-card-text class="pa-0">
        <v-list two-line class="pa-0">
            <template v-for="(item, index) in devices">
                <v-list-item :key="item.uid">
                    <v-list-item-avatar>
                        <v-icon color="success" v-if="item.online">
                            check_circle
                        </v-icon>
                        <v-icon v-else>check_circle</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                        <v-list-item-title v-html="item.uid"></v-list-item-title>
                        <v-list-item-subtitle v-html="item.identity.mac"></v-list-item-subtitle>
                    </v-list-item-content>
                    <v-list-item-action>
                        <TerminalDialog :uid="item.uid" v-if="item.online"></TerminalDialog>
                        <v-btn color="primary" dark icon text @click="remove(item.uid)"><font-awesome-icon icon="trash">trash</font-awesome-icon></v-btn>
                    </v-list-item-action>
                    <v-list-item-action>
                        <v-list-item-action-text v-if="!item.online">last seen {{ item.last_seen | moment("from", "now") }}</v-list-item-action-text>
                    </v-list-item-action>
                </v-list-item>

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
  },

  methods: {
    remove(uid) {
      if (confirm("Are you sure?")) {
        this.$store.dispatch("devices/remove", uid);
      }
    }
  }
};
</script>
