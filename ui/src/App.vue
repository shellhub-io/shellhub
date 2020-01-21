<template>
<v-app>
    <v-navigation-drawer v-if="isLoggedIn" :clipped="clipped" v-model="drawer" :mini-variant="true" enable-resize-watcher app>
        <v-app-bar class="primary" flat>

        </v-app-bar>

        <v-divider></v-divider>
        <v-list>
            <v-list-item v-for="item in items" :key="item.title" :to="item.path" color="secondary">
                <v-list-item-action>
                    <v-icon>{{ item.icon }}</v-icon>
                </v-list-item-action>

                <v-list-item-content>
                    <v-list-item-title>{{ item.title }}</v-list-item-title>
                </v-list-item-content>
            </v-list-item>
        </v-list>
    </v-navigation-drawer>
    <v-app-bar app color="primary lighten-1" flat v-if="isLoggedIn">
        <v-spacer></v-spacer>
        <v-menu transition="scale-transition" origin="top right">
            <template v-slot:activator="{ on }">
                <v-chip v-on="on">
                    <v-icon left>mdi-account</v-icon>
                    {{ $store.getters["auth/currentUser"] }}
                    <v-icon right>mdi-chevron-down</v-icon>
                </v-chip>
            </template>

            <v-card>
                <v-list-item three-line>
                    <v-list-item-content>
                        <v-list-item-title class="mb-1">Tenant ID</v-list-item-title>
                        <v-list-item-subtitle>
                            <v-chip>
                                <span>
                                  {{ tenant }}
                                </span>
                                <v-icon right v-clipboard="tenant" v-clipboard:success="() => { copySnack = true; }">mdi-content-copy</v-icon>
                            </v-chip>
                        </v-list-item-subtitle>
                    </v-list-item-content>
                </v-list-item>

                <v-card-actions>
                    <v-btn small text @click="logout()">
                        Logout
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-menu>
    </v-app-bar>
    <v-content>
        <v-container class="pa-8">
            <router-view></router-view>
        </v-container>
        <v-snackbar v-model="copySnack" :timeout=3000>Tenant ID copied to clipboard</v-snackbar>
    </v-content>
</v-app>
</template>

<script>
export default {
  name: "App",

  methods: {
    logout() {
      this.$store.dispatch("auth/logout").then(() => {
        this.$router.push("/login");
      });
    }
  },

  computed: {
    tenant() {
      return this.$store.getters["auth/tenant"];
    },

    isLoggedIn() {
      return this.$store.getters['auth/isLoggedIn'];
    }
  },

  data() {
    return {
      drawer: true,
      clipped: false,
      copySnack: false,
      items: [
        {
          icon: "dashboard",
          title: "Dashboard",
          path: "/"
        },
        {
          icon: "devices",
          title: "Device Fleet",
          path: "/devices"
        },
        {
          icon: "history",
          title: "Sessions",
          path: "/sessions"
        }
      ]
    };
  }
};
</script>

<style>
.v-list-active {
  border-left: 4px solid var(--v-primary-base);
}

.text-shadow {
  text-shadow: #000 0 0 6px;
  color: transparent;
}
</style>
