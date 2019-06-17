import Vue from "vue";
import Router from "vue-router";
import Dashboard from "./views/Dashboard.vue";

Vue.use(Router);

export default new Router({
    mode: "history",
    base: process.env.BASE_URL,
    routes: [
        {
            path: "/",
            name: "dashboard",
            component: Dashboard
        },
        {
            path: "/devices",
            name: "devices",

            component: () =>
                import(/* webpackChunkName: "devices" */ "./views/Devices.vue")
        },
        {
            path: "/sessions",
            name: "sessions",
            component: () =>
                import("./views/Sessions.vue")
        }
    ]
});