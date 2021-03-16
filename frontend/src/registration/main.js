import Vue from 'vue'
import App from './App.vue'
import 'animate.css'
import 'bootstrap/dist/css/bootstrap.min.css'
import Vuex from 'vuex'

// TODO: could set from environment variable
Vue.config.productionTip = false

Vue.use(Vuex)

const store = new Vuex.Store({
  state: {
    data: {
      apiVersion: process.env.VUE_APP_API_VERSION,
      create: {
        id: "",
        password: "",
        show: false,
      },
      join: {
        id: "",
        password: "",
        show: false,
      },
      alert: {
        message: "",
      },
    },
  },
  mutations: {
    setAlertMessage: function (state, message) {
      state.data.alert.message = message
    },
    showCreateRoom: function (state) {
      let newState = JSON.parse(JSON.stringify(state))
      newState.data.create.show = true
      newState.data.join.show = false
      state.data = newState.data
    },
    showJoinRoom: function (state) {
      let newState = JSON.parse(JSON.stringify(state))
      newState.data.create.show = false
      newState.data.join.show = true
      state.data = newState.data
    },
    setId: function (state, payload) {
      let { id, type } = payload
      state.data[type].id = id
    },
    setPassword: function (state, payload) {
      let { password, type } = payload
      state.data[type].password = password
    }
  },
  actions: {
    submit: function (context, type) {
      let path = './api/' + context.state.data.apiVersion + '/registration/' + type
      fetch(path, {
        method: "post",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          id: context.state.data[type].id,
          password: context.state.data[type].password,
        }),
      })
        .then((response) => {
          if (response.ok) {
            context.commit("setAlertMessage", "");
            let roomURL = response.headers.get("Content-Location");
            window.location = roomURL;
          } else {
            response.text().then((text) => {
              let message = text ? text : "internal server error";
              context.commit("setAlertMessage", message)
            });
          }
        })
        .catch((error) => {
          let message = error.message ? error.message : "internal server error";
          context.commit("setAlertMessage", message)
        });
    }
  },
  getters: {
    alert: function (state) {
      return state.data.alert
    },
    create: function (state) {
      return state.data.create
    },
    join: function (state) {
      return state.data.join
    }
  }
})

new Vue({
  render: h => h(App),
  store: store,
}).$mount('#app')